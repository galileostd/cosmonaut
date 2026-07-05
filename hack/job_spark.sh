#!/bin/bash
# run-spark-job.sh - Executes a Spark Pi job in Cosmonaut

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     🚀 RUNNING SPARK JOB IN COSMONAUT                  ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"

# Configuration
NAMESPACE="cosmonaut"
JOB_NAME="spark-pi-$(date +%s)"  # unique name
JOB_LABEL="spark-pi"
JOB_GROUP="examples-math"
COMPONENT_NAME="spark"
COMPONENT_NAMESPACE="cosmonaut"
COSMONAUT_API="http://localhost:8080"

echo -e "\n${YELLOW}📋 Configuration:${NC}"
echo "  - Physical Job Name: $JOB_NAME"
echo "  - Logical Name: $JOB_LABEL"
echo "  - Group: $JOB_GROUP"
echo "  - Namespace: $NAMESPACE"

# Check if port-forward is running
echo -e "\n${YELLOW}🔍 Checking Cosmonaut connectivity...${NC}"
if ! curl -s -o /dev/null -w "%{http_code}" "$COSMONAUT_API/api/v1/system/health" | grep -q "200"; then
    echo -e "${RED}❌ Cosmonaut not accessible at $COSMONAUT_API${NC}"
    echo -e "${YELLOW}💡 Run in another terminal:${NC}"
    echo "  kubectl port-forward -n cosmonaut svc/cosmonaut 8080:8080"
    exit 1
fi
echo -e "${GREEN}✅ Cosmonaut is responding${NC}"

# Check if Spark plugin is registered
echo -e "\n${YELLOW}🔍 Checking Spark plugin...${NC}"
SPARK_PLUGIN=$(curl -s "$COSMONAUT_API/api/v1/plugins" | jq -r '.items[] | select(.name=="spark") | .name')
if [ -z "$SPARK_PLUGIN" ]; then
    echo -e "${RED}❌ Spark plugin not found${NC}"
    echo "Available plugins:"
    curl -s "$COSMONAUT_API/api/v1/plugins" | jq '.items[].name'
    exit 1
fi
echo -e "${GREEN}✅ Spark plugin found: $SPARK_PLUGIN${NC}"

# Check if Spark component exists
echo -e "\n${YELLOW}🔍 Checking Spark component...${NC}"
COMPONENT=$(curl -s "$COSMONAUT_API/api/v1/components" | jq -r ".items[] | select(.metadata.name==\"$COMPONENT_NAME\") | .metadata.name")
if [ -z "$COMPONENT" ]; then
    echo -e "${RED}❌ Spark component not found${NC}"
    echo "Available components:"
    curl -s "$COSMONAUT_API/api/v1/components" | jq '.items[].metadata.name'
    exit 1
fi
echo -e "${GREEN}✅ Spark component found${NC}"

# Method 1: Submit via kubectl (recommended)
echo -e "\n${YELLOW}📦 Method 1: Submitting via kubectl (recommended)${NC}"

# Create the SparkApplication YAML
cat > /tmp/spark-job.yaml <<EOF
apiVersion: sparkoperator.k8s.io/v1beta2
kind: SparkApplication
metadata:
  name: $JOB_NAME
  namespace: $NAMESPACE
  labels:
    cosmonaut.galileostd.io/job-name: "$JOB_LABEL"
    cosmonaut.galileostd.io/job-group: "$JOB_GROUP"
spec:
  type: Scala
  mode: cluster
  image: "bitnami/spark:latest"
  imagePullPolicy: IfNotPresent
  mainClass: org.apache.spark.examples.SparkPi
  mainApplicationFile: "local:///opt/bitnami/spark/examples/jars/spark-examples_2.12-3.4.1.jar"
  sparkVersion: "3.4.1"
  restartPolicy:
    type: Never
  driver:
    cores: 1
    memory: "512m"
    serviceAccount: spark
  executor:
    cores: 1
    instances: 2
    memory: "512m"
EOF

echo -e "${YELLOW}Applying SparkApplication CRD...${NC}"
kubectl apply -f /tmp/spark-job.yaml

echo -e "${GREEN}✅ Job submitted via kubectl${NC}"
echo "  - kubectl get sparkapp -n $NAMESPACE $JOB_NAME"

# Method 2: Submit via Cosmonaut API (alternative)
echo -e "\n${YELLOW}📡 Method 2: Submitting via Cosmonaut API (alternative)${NC}"
echo -e "${YELLOW}💡 Uncomment the curl command below to try this method${NC}"

# Uncomment to try the API method:
# echo -e "\n${YELLOW}Submitting via Cosmonaut API...${NC}"
# curl -X POST "$COSMONAUT_API/api/v1/components/$COMPONENT_NAMESPACE/$COMPONENT_NAME/exec" \
#   -H "Content-Type: application/json" \
#   -d "{
#     \"job_name\": \"$JOB_LABEL\",
#     \"job_group\": \"$JOB_GROUP\",
#     \"spec\": {
#       \"type\": \"Scala\",
#       \"mode\": \"cluster\",
#       \"image\": \"bitnami/spark:latest\",
#       \"mainClass\": \"org.apache.spark.examples.SparkPi\",
#       \"mainApplicationFile\": \"local:///opt/bitnami/spark/examples/jars/spark-examples_2.12-3.4.1.jar\",
#       \"sparkVersion\": \"3.4.1\",
#       \"restartPolicy\": { \"type\": \"Never\" },
#       \"driver\": { \"cores\": 1, \"memory\": \"512m\", \"serviceAccount\": \"spark\" },
#       \"executor\": { \"cores\": 1, \"instances\": 2, \"memory\": \"512m\" }
#     }
#   }" | jq '.'

# Monitor the job
echo -e "\n${YELLOW}👀 Monitoring job in Cosmonaut...${NC}"
echo -e "${BLUE}Press Ctrl+C to stop monitoring (job continues running)${NC}"

# Function to check job status via Cosmonaut API
check_job_status() {
    echo -e "\n${YELLOW}📊 Job Status (via Cosmonaut API):${NC}"
    curl -s "$COSMONAUT_API/api/v1/jobs?job_name=$JOB_LABEL" | jq -r '.items[] | {
        id: .job_id,
        name: .job_name,
        group: .job_group,
        state: .state,
        message: .message
    } | "  - ID: \(.id)\n    Name: \(.name)\n    Group: \(.group)\n    State: \(.state)\n    Message: \(.message)\n"'
}

# Initial status
check_job_status

# Monitor with kubectl as well
echo -e "\n${YELLOW}📊 Job Status (via kubectl):${NC}"
kubectl get sparkapp -n $NAMESPACE $JOB_NAME -o wide

# Provide commands for further inspection
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Job submitted successfully!${NC}"
echo -e "\n${YELLOW}📌 Useful commands for inspection:${NC}"

# Wait a bit for the job to start
echo -e "\n${YELLOW}⏳ Waiting 5 seconds for job to initialize...${NC}"
sleep 5

# Show job details
echo -e "\n${YELLOW}📊 Checking job status in Cosmonaut UI:${NC}"
echo "  → http://localhost:8080/jobs?job_name=$JOB_LABEL"

echo -e "\n${YELLOW}🔍 kubectl commands:${NC}"
echo "  # See all Spark jobs:"
echo "  kubectl get sparkapp -n $NAMESPACE"
echo ""
echo "  # See job details:"
echo "  kubectl describe sparkapp $JOB_NAME -n $NAMESPACE"
echo ""
echo "  # Get driver logs (after job starts):"
echo "  kubectl logs -n $NAMESPACE -l spark-role=driver,spark-app-name=$JOB_NAME"
echo ""
echo "  # Get executor logs:"
echo "  kubectl logs -n $NAMESPACE -l spark-role=executor,spark-app-name=$JOB_NAME"
echo ""
echo "  # Delete the job:"
echo "  kubectl delete sparkapp $JOB_NAME -n $NAMESPACE"

echo -e "\n${YELLOW}🌐 Cosmonaut API endpoints:${NC}"
echo "  # List all jobs:"
echo "  curl -s $COSMONAUT_API/api/v1/jobs | jq"
echo ""
echo "  # Filter by job name:"
echo "  curl -s '$COSMONAUT_API/api/v1/jobs?job_name=$JOB_LABEL' | jq"
echo ""
echo "  # Get specific job details:"
echo "  curl -s $COSMONAUT_API/api/v1/jobs/\$JOB_ID | jq"
echo ""
echo "  # Get service logs:"
echo "  curl -s '$COSMONAUT_API/api/v1/components/$COMPONENT_NAMESPACE/$COMPONENT_NAME/logs?tail=50' | jq"

# Continuous monitoring option
echo -e "\n${YELLOW}🔄 Continuous monitoring? (y/n)${NC}"
read -t 5 -p "> " MONITOR || MONITOR="n"

if [[ $MONITOR == "y" || $MONITOR == "Y" ]]; then
    echo -e "\n${BLUE}Starting continuous monitoring (Ctrl+C to stop)...${NC}"
    while true; do
        clear
        echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
        echo -e "${BLUE}  🚀 SPARK JOB MONITOR - $(date)${NC}"
        echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
        
        echo -e "\n${YELLOW}📊 Job Status:${NC}"
        curl -s "$COSMONAUT_API/api/v1/jobs?job_name=$JOB_LABEL" | jq -r '.items[] | {
            id: .job_id,
            name: .job_name,
            state: .state,
            message: .message,
            details: .details
        }'
        
        echo -e "\n${YELLOW}📦 Kubernetes Status:${NC}"
        kubectl get sparkapp -n $NAMESPACE $JOB_NAME -o wide 2>/dev/null || echo "Job not found (may be deleted)"
        
        echo -e "\n${YELLOW}📋 Recent Logs:${NC}"
        kubectl logs -n $NAMESPACE -l spark-role=driver,spark-app-name=$JOB_NAME --tail=10 2>/dev/null || echo "No logs yet (driver may not be running)"
        
        echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}⏳ Refreshing in 5 seconds... (Ctrl+C to stop)${NC}"
        sleep 5
    done
fi

echo -e "\n${GREEN}✅ Done!${NC}"