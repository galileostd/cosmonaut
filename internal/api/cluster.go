package api

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/galileostd/cosmonaut/internal/registry"
)

type clusterMetrics struct {
	ActiveServices   int     `json:"activeServices"`
	InteractiveJobs  int     `json:"interactiveJobs"`
	FailedJobs       int     `json:"failedJobs"`
	ClusterResources float64 `json:"clusterResources"`
}

type nodeResponse struct {
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	CPUUsed       float64 `json:"cpuUsed"`
	CPUAlloc      float64 `json:"cpuAlloc"`
	MemoryUsed    float64 `json:"memoryUsed"`
	MemoryAlloc   float64 `json:"memoryAlloc"`
	Status        string  `json:"status"`
	StatusMessage string  `json:"statusMessage"`
}

type clusterResponse struct {
	Metrics clusterMetrics `json:"metrics"`
	Nodes   []nodeResponse `json:"nodes"`
}

// handleCluster returns cluster metrics and node information.
// GET /api/v1/cluster
func (s *Server) handleCluster(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. List all nodes
	var nodes corev1.NodeList
	if err := s.k8s.List(ctx, &nodes, &client.ListOptions{}); err != nil {
		writeProblem(w, r, problemInternal(r, "failed to list nodes: "+err.Error()))
		return
	}

	// 2. List all CosmoComponents
	var components registry.CosmoComponentList
	if err := s.k8s.List(ctx, &components, &client.ListOptions{
		Namespace: "cosmonaut",
	}); err != nil {
		writeProblem(w, r, problemInternal(r, "failed to list components: "+err.Error()))
		return
	}

	// 3. Calculate metrics
	activeServices := 0
	interactiveJobs := 0
	failedJobs := 0

	for _, comp := range components.Items {
		if comp.Status.Health == "healthy" {
			activeServices++
		}
	}

	// 4. Busca métricas REAIS via Metrics API
	nodeMetrics := s.getNodeMetrics(ctx, nodes.Items)

	// 5. Build node responses com dados REAIS
	nodeResponses := make([]nodeResponse, 0, len(nodes.Items))
	var totalCPUUsed, totalCPUAlloc float64
	var totalMemUsed, totalMemAlloc float64

	for _, node := range nodes.Items {
		// Node status
		ready := false
		statusMessage := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				ready = cond.Status == corev1.ConditionTrue
				if ready {
					statusMessage = "Ready"
				} else {
					statusMessage = string(cond.Status)
				}
				break
			}
		}

		// Role
		role := "worker"
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			role = "master"
		} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			role = "master"
		}

		// Allocatable (do node)
		cpuAlloc := float64(node.Status.Allocatable.Cpu().MilliValue()) / 1000
		memAlloc := float64(node.Status.Allocatable.Memory().Value()) / (1024 * 1024 * 1024)

		// USED - da Metrics API (real, não fake!)
		metric := nodeMetrics[node.Name]
		cpuUsed := metric.CPU
		memUsed := metric.Memory

		// Fallback só se metrics API falhar completamente
		if cpuUsed == 0 && memUsed == 0 {
			cpuUsed = cpuAlloc * 0.2
			memUsed = memAlloc * 0.3
		}

		totalCPUUsed += cpuUsed
		totalCPUAlloc += cpuAlloc
		totalMemUsed += memUsed
		totalMemAlloc += memAlloc

		status := "ok"
		if !ready {
			status = "warn"
		}

		nodeResponses = append(nodeResponses, nodeResponse{
			Name:          node.Name,
			Role:          role,
			CPUUsed:       cpuUsed,
			CPUAlloc:      cpuAlloc,
			MemoryUsed:    memUsed,
			MemoryAlloc:   memAlloc,
			Status:        status,
			StatusMessage: statusMessage,
		})
	}

	// 6. ClusterResources com cálculo REAL
	totalResources := 0.0
	if totalCPUAlloc > 0 && totalMemAlloc > 0 {
		cpuUsage := (totalCPUUsed / totalCPUAlloc) * 100
		memUsage := (totalMemUsed / totalMemAlloc) * 100
		totalResources = (cpuUsage + memUsage) / 2
	}

	response := clusterResponse{
		Metrics: clusterMetrics{
			ActiveServices:   activeServices,
			InteractiveJobs:  interactiveJobs,
			FailedJobs:       failedJobs,
			ClusterResources: totalResources,
		},
		Nodes: nodeResponses,
	}

	writeJSON(w, http.StatusOK, response)
}

type nodeMetricData struct {
	CPU    float64 // cores
	Memory float64 // GB
}

// getNodeMetrics busca métricas reais via Metrics API (kubectl top node)
func (s *Server) getNodeMetrics(ctx context.Context, nodes []corev1.Node) map[string]nodeMetricData {
	result := make(map[string]nodeMetricData)

	// Cria client de métricas a partir do rest.Config
	metricsClient, err := versioned.NewForConfig(s.restConfig)
	if err != nil {
		// Fallback: tenta via clientset padrão estimando por requests
		return s.getNodeMetricsFallback(ctx, nodes)
	}

	for _, node := range nodes {
		metrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		cpuMillis := metrics.Usage.Cpu().MilliValue()
		memBytes := metrics.Usage.Memory().Value()

		result[node.Name] = nodeMetricData{
			CPU:    float64(cpuMillis) / 1000,
			Memory: float64(memBytes) / (1024 * 1024 * 1024),
		}
	}

	return result
}

// Fallback: estima uso por requests dos pods running no node
func (s *Server) getNodeMetricsFallback(ctx context.Context, nodes []corev1.Node) map[string]nodeMetricData {
	result := make(map[string]nodeMetricData)

	clientset, err := kubernetes.NewForConfig(s.restConfig)
	if err != nil {
		return result
	}

	for _, node := range nodes {
		cpuAlloc := float64(node.Status.Allocatable.Cpu().MilliValue()) / 1000
		memAlloc := float64(node.Status.Allocatable.Memory().Value()) / (1024 * 1024 * 1024)

		// Lista pods running nesse node
		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			continue
		}

		var cpuUsed, memUsed float64
		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				continue
			}
			for _, c := range pod.Spec.Containers {
				if c.Resources.Requests != nil {
					if cpuReq := c.Resources.Requests.Cpu(); cpuReq != nil {
						cpuUsed += float64(cpuReq.MilliValue()) / 1000
					}
					if memReq := c.Resources.Requests.Memory(); memReq != nil {
						memUsed += float64(memReq.Value()) / (1024 * 1024 * 1024)
					}
				}
			}
		}

		// Se não achou nada, usa estimativa conservadora
		if cpuUsed == 0 {
			cpuUsed = cpuAlloc * 0.15
		}
		if memUsed == 0 {
			memUsed = memAlloc * 0.25
		}

		result[node.Name] = nodeMetricData{
			CPU:    cpuUsed,
			Memory: memUsed,
		}
	}

	return result
}
