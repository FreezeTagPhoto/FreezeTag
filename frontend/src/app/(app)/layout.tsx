import SidebarShell from "@/components/Sidebar/SidebarShell";

export default function AppLayout({ children }: { children: React.ReactNode }) {
    return <SidebarShell>{children}</SidebarShell>;
}
