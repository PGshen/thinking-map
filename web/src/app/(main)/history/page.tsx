"use client";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { fetchMapList } from "@/api/map";
import type { Map } from "@/types/map";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardHeader, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { AlertDialogHeader, AlertDialogFooter, AlertDialog, AlertDialogTrigger, AlertDialogContent, AlertDialogTitle, AlertDialogDescription, AlertDialogCancel, AlertDialogAction } from "@/components/ui/alert-dialog";

export default function Page() {
  // Placeholder state for future data
  const [maps, setMaps] = useState<Map[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [total, setTotal] = useState(0);
  const limit = 9;
  const router = useRouter();
  const [problemType, setProblemType] = useState("全部类型");
  const [dateRange, setDateRange] = useState("this-month");
  const [search, setSearch] = useState("");

  const fetchData = async (reset = false) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetchMapList({
        page: reset ? 1 : page,
        limit,
        problemType: problemType !== "全部类型" ? problemType : undefined,
        dateRange: dateRange,
        search: search || undefined,
      });
      if (res && res.data) {
        setTotal(res.data.total);
        setHasMore(res.data.page * res.data.limit < res.data.total);
        setMaps(reset ? res.data.items : prev => [...prev, ...res.data.items]);
        setPage(res.data.page);
      } else {
        setError("未能获取历史记录");
      }
    } catch (e: any) {
      setError(e.message || "请求失败");
      toast.error(e.message || "请求失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData(true);
  }, [problemType, search, dateRange]);

  const handleLoadMore = () => {
    setPage(p => p + 1);
  };

  useEffect(() => {
    if (page > 1) fetchData();
  }, [page]);

  const handleView = (id: string) => {
    router.push(`/map/${id}`);
  };

  return (
    <div className="w-full mx-auto px-3 py-2 min-h-screen">
      <div className="sticky top-0 bg-background z-10 pt-4 pb-4 border-b">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <Select value={problemType} onValueChange={setProblemType}>
              <SelectTrigger className="w-[120px]">
                <SelectValue placeholder="全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="全部类型">全部</SelectItem>
                <SelectItem value="研究型">研究型</SelectItem>
                <SelectItem value="创意型">创意型</SelectItem>
                <SelectItem value="分析型">分析型</SelectItem>
                <SelectItem value="规划型">规划型</SelectItem>
              </SelectContent>
            </Select>

            <Select value={dateRange} onValueChange={setDateRange}>
              <SelectTrigger className="w-[120px]">
                <SelectValue placeholder="本周" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="this-week">本周</SelectItem>
                <SelectItem value="last-week">上周</SelectItem>
                <SelectItem value="this-month">本月</SelectItem>
                <SelectItem value="all-time">所有时间</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex-1 max-w-sm">
            <Input
              type="search"
              placeholder="搜索历史记录..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>
      </div>
      {/* 历史卡片区 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-2 mt-4">
        {maps?.map((map) => (
          <Card key={map.id} className="gap-0! py-4!">
            <CardHeader className="px-4 py-1">
              <h3 className="font-semibold mb-2">{map.problem}</h3>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-2">{map.target}</p>
              <div className="flex items-center gap-2 mb-2">
                <Badge variant="secondary">{map.problemType}</Badge>
                <Badge variant="outline">{`完成度: ${map.metadata?.progress|0}%`}</Badge>
              </div>
              <div className="text-sm text-muted-foreground">
                {map.createdAt}
              </div>
            </CardContent>
            <CardFooter className="flex flex-wrap gap-2">
              <Button size="sm" variant="default" className="cursor-pointer" onClick={() => handleView(map.id)}>继续</Button>
              <Button size="sm" variant="outline" className="cursor-pointer">导出</Button>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button size="sm" variant="destructive" className="cursor-pointer">删除</Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>确认删除</AlertDialogTitle>
                    <AlertDialogDescription>
                      确定要删除这条会话记录吗？此操作无法撤销。
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>取消</AlertDialogCancel>
                    <AlertDialogAction>
                      确认
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </CardFooter>
          </Card>
        ))}
        {loading && (
          <div className="col-span-3 text-center text-gray-400 py-20">加载中...</div>
        )}
        {error && (
          <div className="col-span-3 text-center text-red-400 py-20">{error}</div>
        )}
        {!loading && !error && maps.length === 0 && (
          <div className="col-span-3 text-center text-gray-400 py-20">暂无历史记录</div>
        )}
      </div>
      {hasMore && !loading && !error && (
        <div className="flex justify-center mt-8">
          <Button variant="outline" onClick={handleLoadMore}>加载更多...</Button>
        </div>
      )}
    </div>
  );
} 