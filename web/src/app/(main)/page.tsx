"use client";
import Image from "next/image";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { useState } from "react";
import { Command } from "lucide-react";
import { useGlobalStore } from "@/store/globalStore";


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
    </div>
  );
}
