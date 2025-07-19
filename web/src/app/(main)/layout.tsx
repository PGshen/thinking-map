import { AuthGuard } from '@/components/auth-guard';
import SidebarLayout from '@/layouts/sidebar-layout';

export default function SiderLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <SidebarLayout>
        {children}
      </SidebarLayout>
    </AuthGuard>
  );
}
