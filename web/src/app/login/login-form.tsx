"use client"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useState } from "react"
import { loginUser } from "@/api/auth"
import { useGlobalStore } from "@/store/globalStore"
import { useRouter } from "next/navigation"
import { setToken } from "@/lib/auth"
import { toast } from "sonner"
import { Eye, EyeOff } from "lucide-react"

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const setUser = useGlobalStore((s) => s.setUser);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!email || !password) {
      setError("请输入邮箱和密码");
      return;
    }
    setLoading(true);
    try {
      const response = await loginUser({ email, password });
      if (response.code === 200 && response.data) {
        const data = response.data;
        if (typeof window !== 'undefined') {
          setToken(data.accessToken || '', data.refreshToken || '');
        }
        setUser({ userId: data.userId, username: data.username, email: data.email, fullName: data.fullName });
        toast.success("登录成功！")
        router.push("/");
      } else {
        toast.error(response.message || "登录失败")
        setError(response.message || "登录失败");
      }
    } catch (err: any) {
      setError(err?.message || "登录失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>欢迎登录</CardTitle>
          <CardDescription>
            输入邮箱登录账号
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <div className="flex flex-col gap-6">
              <div className="grid gap-3">
                <Label htmlFor="email">邮箱</Label>
                <Input
                  id="email"
                  type="text"
                  placeholder="邮箱"
                  required
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                />
              </div>
              <div className="grid gap-3">
                <div className="flex items-center">
                  <Label htmlFor="password">密码</Label>
                  <a
                    href="#"
                    className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                  >
                    忘记密码?
                  </a>
                </div>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex flex-col gap-3">
                <Button type="submit" className="w-full cursor-pointer" disabled={loading}>
                  {loading ? "登录中..." : "登录"}
                </Button>
              </div>
            </div>
            <div className="mt-4 text-center text-sm">
              没有帐号?{" "}
              <a href="/register" className="underline underline-offset-4">
                注册
              </a>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
