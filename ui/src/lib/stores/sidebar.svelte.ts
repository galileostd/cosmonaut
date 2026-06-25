class SidebarStore {
	collapsed = $state(false);
	get width() { return this.collapsed ? 52 : 220; }
}

export const sidebar = new SidebarStore();




