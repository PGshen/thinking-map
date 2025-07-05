import SidebarLayout from '@/layouts/sidebar-layout';

export default function SiderLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarLayout>
      {children}
    </SidebarLayout>
  );
}
