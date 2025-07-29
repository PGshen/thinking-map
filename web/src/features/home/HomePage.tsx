"use client";
import Image from "next/image";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { useState } from "react";
import { Command, Loader } from "lucide-react";
import { useGlobalStore } from "@/store/globalStore";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { EditableTextarea } from "@/components/ui/editable-textarea";
import { toast } from "sonner";
import { SseJsonStreamParser } from "@/lib/sse-parser";
import API_ENDPOINTS from "@/api/endpoints";
import { useRouter } from "next/navigation";
import { createMap } from "@/api/map";
import type { CreateMapRequest } from "@/types/map";

const PROBLEM_TYPES = [
  { value: '研究型', label: '研究型', description: '深入探索和分析特定主题' },
  { value: '创意型', label: '创意型', description: '发散思维，寻找创新解决方案' },
  { value: '分析型', label: '分析型', description: '系统分析数据和现象' },
  { value: '规划型', label: '规划型', description: '制定策略和执行计划' }
];

const PROBLEM_EXAMPLES = {
  研究型: '如何评估人工智能在教育领域的应用效果？',
  创意型: '设计一个创新的城市共享单车系统',
  分析型: '分析全球供应链中断对电子产品市场的影响',
  规划型: '制定一个为期6个月的产品推广策略'
};

export default function HomePage() {
  const [problem, setProblem] = useState('');
  const [problemType, setProblemType] = useState('研究型');
  const user = useGlobalStore((s) => s.user);
  const isLoggedIn = !!user;

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [goal, setGoal] = useState('');
  const [keyPoints, setKeyPoints] = useState<string[]>([]);
  const [constraints, setConstraints] = useState<string[]>([]);
  const [suggestion, setSuggestion] = useState('');
  const [supplementary, setSupplementary] = useState(''); // 补充消息
  const [parentMsgID, setParentMsgID] = useState('');

  const [isLoading, setIsLoading] = useState(false);

  const router = useRouter();

  const handleUnderstanding = () => {
    if (supplementary.length > 0) {
      // 填写了补充信息
      setProblem('')
    } else if (problem.length < 6) {
      toast.info("问题描述至少需要6个字，请补充更多细节。");
      return;
    }
    // 打开问题理解确认抽屉
    setIsDialogOpen(true);
    setIsLoading(true);

    let url = API_ENDPOINTS.THINKING.UNDERSTANDING
    const jsonStream = new SseJsonStreamParser(true)
    jsonStream.on("title", (data) => {
      setTitle(data)
      setIsLoading(false)
    })
    jsonStream.on("problem", (data) => {
      setProblem(data)
      setIsLoading(false)
    })
    jsonStream.on("goal", (data) => {
      setGoal(data)
      setIsLoading(false)
    })
    jsonStream.on("keyPoints[*]", (data, path) => {
      const index = parseInt(path[path.length - 1] as string)
      setKeyPoints(prev => {
        const newKeyPoints = [...(prev || [])]
        newKeyPoints[index] = data
        return newKeyPoints
      })
      setIsLoading(false)
    })
    jsonStream.on("constraints[*]", (data, path) => {
      const index = parseInt(path[path.length - 1] as string)
      setConstraints(prev => {
        const newConstraints = [...(prev || [])]
        newConstraints[index] = data
        return newConstraints
      })
      setIsLoading(false)
    })
    jsonStream.on("suggestion", (data) => {
      setSuggestion(data)
      setIsLoading(false)
    })
    jsonStream.connect("POST", url, {
      parentMsgID: parentMsgID,
      problem: problem,
      problemType: problemType,
      supplementary: supplementary,
    }, (ev) => {
      if (ev.event == "id") {
        setParentMsgID(ev.data)
      }
    })
  };

  const handleCreateMap = async () => {
    if (!problem.trim()) {
      toast.error("问题描述不能为空");
      return;
    }
    const params: CreateMapRequest = {
      title: title.trim(),
      problem: problem.trim(),
      problemType: problemType || undefined,
      target: goal.trim() || undefined,
      keyPoints: keyPoints.length > 0 ? keyPoints : undefined,
      constraints: constraints.length > 0 ? constraints : undefined,
    };
    try {
      const res = await createMap(params);
      if (res && res.data && res.data.id) {
        setIsDialogOpen(false);
        router.push(`/map/${res.data.id}`);
      } else {
        toast.error("创建思维导图失败");
      }
    } catch (e) {
      toast.error("创建思维导图失败");
    }
  };

  return (
    <div className="container mx-auto px-4 py-8 flex flex-col items-center justify-center min-h-screen">
      {/* 品牌展示区 */}
      <div className="text-center mb-8">
        <div className="flex mb-2 items-end ">
          <div className="mr-4 flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg">
            <Command className="size-6" />
          </div>
          {isLoggedIn ? (
            <h1 className="text-4xl font-bold">Welcome, {user?.username}</h1>
          ) : (
            <h1 className="text-4xl font-bold">Welcome</h1>
          )}
        </div>
        <p className="text-lg text-gray-600">可视化解构您的思维历程，让问题解决变得透明和可控</p>
      </div>

      {/* 问题输入区 */}
      <div className="w-full max-w-2xl space-y-3">
        <div className="relative">
          <Textarea
            placeholder="请输入您的问题（50-200字最佳）..."
            value={problem}
            onChange={(e) => setProblem(e.target.value)}
            className="min-h-[120px] pr-24"
          />
          <div className="absolute bottom-2 right-2 flex items-center space-x-2">
            <span className="text-sm text-gray-500">
              {problem.length} / 200 字
              {problem.length > 200 && (
                <span className="text-yellow-500 ml-2">字数超出建议范围</span>
              )}
            </span>
            <Button
              onClick={handleUnderstanding}
              disabled={!problem.trim() || !problemType}
              className="ml-2"
              size="sm"
            >
              提交问题
            </Button>
          </div>
        </div>

        <div className="space-y-3">
          <Label className="text-base">选择问题类型</Label>
          <RadioGroup
            value={problemType}
            onValueChange={setProblemType}
            className="grid grid-cols-1 md:grid-cols-2 gap-1!"
          >
            {PROBLEM_TYPES.map((type) => (
              <div key={type.value} className="flex items-center space-x-2 p-1 rounded-lg hover:bg-slate-50">
                <RadioGroupItem value={type.value} id={type.value} />
                <Label htmlFor={type.value} className="flex-1 cursor-pointer">
                  <span className="font-medium">{type.label}</span>
                  <span className="text-gray-500 ml-2">- {type.description}</span>
                </Label>
              </div>
            ))}
          </RadioGroup>
        </div>

        {problemType && (
          <div className="text-sm text-gray-600">
            <p>示例问题：{PROBLEM_EXAMPLES[problemType as keyof typeof PROBLEM_EXAMPLES]}</p>
          </div>
        )}
      </div>

      {/* 问题理解确认抽屉 */}
      <Dialog open={isDialogOpen} onOpenChange={(open) => {
          // 只有点击取消按钮时才允许关闭对话框
          if (!open) return;
          setIsDialogOpen(open);
        }}>
          <DialogContent className="max-w-2xl!">
            <DialogHeader>
              <DialogTitle>问题理解确认</DialogTitle>
              <DialogDescription>请确认系统对问题的理解是否准确</DialogDescription>
            </DialogHeader>

            <div className="max-h-[60vh] overflow-y-auto space-y-4 py-4">
              {isLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
                  <span className="ml-2">正在分析问题...</span>
                </div>
              ) : (
                <>
                  <div className="space-y-2">
                    <h3 className="font-medium">标题</h3>
                    <EditableTextarea
                      value={title}
                      onChange={setTitle}
                      className="h-[40px]"
                      showDeleteButton={false}
                    />
                  </div>
                  <div className="space-y-2">
                    <h3 className="font-medium">系统理解</h3>
                    <EditableTextarea
                      value={problem}
                      onChange={setProblem}
                      className="h-[40px]"
                      showDeleteButton={false}
                    />
                  </div>

                  {
                    goal != "" && <div className="space-y-2 group relative">
                      <h3 className="font-medium">问题目标</h3>
                      <EditableTextarea
                        value={goal}
                        onChange={setGoal}
                        className="h-[40px]"
                        showDeleteButton={false}
                      />
                    </div>
                  }

                  {
                    keyPoints.length > 0 && <div className="space-y-2">
                      <div className="flex justify-between items-center">
                        <h3 className="font-medium">核心要点</h3>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => setKeyPoints([...keyPoints, ''])}
                        >
                          添加要点
                        </Button>
                      </div>
                      <div className="space-y-2">
                        {keyPoints.map((point, index) => (
                          <EditableTextarea
                            key={index}
                            value={point}
                            onChange={(value) => {
                              const newPoints = [...keyPoints];
                              newPoints[index] = value;
                              setKeyPoints(newPoints);
                            }}
                            onDelete={() => {
                              const newPoints = keyPoints.filter((_, i) => i !== index);
                              setKeyPoints(newPoints);
                            }}
                          />
                        ))}
                      </div>
                    </div>
                  }

                  {
                    constraints.length > 0 && <div className="space-y-2">
                      <div className="flex justify-between items-center">
                        <h3 className="font-medium">目标约束</h3>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => setConstraints([...constraints, ''])}
                        >
                          添加约束
                        </Button>
                      </div>
                      <div className="space-y-2">
                        {constraints.map((constraint, index) => (
                          <EditableTextarea
                            key={index}
                            value={constraint}
                            onChange={(value) => {
                              const newConstraints = [...constraints];
                              newConstraints[index] = value;
                              setConstraints(newConstraints);
                            }}
                            onDelete={() => {
                              const newConstraints = constraints.filter((_, i) => i !== index);
                              setConstraints(newConstraints);
                            }}
                          />
                        ))}
                      </div>
                    </div>
                  }

                  {
                    suggestion != "" && <div className="space-y-2">
                      <h3 className="font-medium">问题建议</h3>
                      <p className='text-red-400 text-sm'>{suggestion}</p>
                    </div>
                  }

                  <div className='space-y-2'>
                    <h3 className="font-medium text-red-400">我要补充信息</h3>
                    <Textarea
                      placeholder="补充信息（可选）..."
                      value={supplementary}
                      onChange={(e) => setSupplementary(e.target.value)}
                      className="min-h-[80px]"
                    />
                  </div>
                </>
              )}
            </div>

            <DialogFooter className="flex-row justify-end space-x-2">
              <Button variant="outline" onClick={() => {
                // 点击取消按钮时，清空所有状态数据
                setTitle('');
                setProblem('');
                setProblemType('');
                setKeyPoints([]);
                setGoal('');
                setConstraints([]);
                setSuggestion('');
                setSupplementary('');
                setIsDialogOpen(false);
              }}>
                取消
              </Button>
              <Button variant="outline" onClick={handleUnderstanding}>
                需要调整
              </Button>
              <Button onClick={handleCreateMap} disabled={isLoading}>理解正确，开始分析</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
    </div>
  );
} 