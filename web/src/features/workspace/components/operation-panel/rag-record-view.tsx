import { RAGRecord } from "@/types/message";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
// Removed Card components in favor of div with custom styles
import { Link, Search, ChevronDown, ChevronRight, FileCheck } from "lucide-react";
import { useState } from "react";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";

interface RAGRecordViewProps {
  ragRecord: RAGRecord;
}

export function RAGRecordView({ ragRecord }: RAGRecordViewProps) {
  const [isFollowUpOpen, setIsFollowUpOpen] = useState(true);

  return (
    <div className="mt-4 space-y-4 w-full">
      <div>
        <h4 className="mb-2 flex items-center gap-2 text-sm font-medium text-purple-700">
          <FileCheck className="h-4 w-4" />
          <span>结果总结</span>
        </h4>
        <p className="text-sm text-gray-800">{ragRecord.answer}</p>
      </div>

      {ragRecord.results && ragRecord.results.length > 0 && (
        <div>
          <h4 className="mb-2 flex items-center gap-2 text-sm font-medium text-purple-700">
            <Link className="h-4 w-4" />
            <span>引用来源</span>
          </h4>
          <div className="w-full">
            <div className="grid gap-4 pb-2 pr-2 grid-cols-[repeat(auto-fit,minmax(16rem,1fr))]">
              {ragRecord.results.map((result, index) => (
                <HoverCard key={index}>
                  <HoverCardTrigger asChild>
                    <a href={result.url} target="_blank" rel="noopener noreferrer" className="block">
                      <div className="w-full cursor-pointer transition-shadow duration-300 hover:shadow-lg rounded-lg border border-purple-200 bg-white">
                        <div className="p-3">
                          <div className="flex items-center gap-2">
                            {result.favicon && ((result as any).favicon || (result.url && `https://www.google.com/s2/favicons?sz=32&domain_url=${encodeURIComponent(result.url)}`)) && (
                              <img
                                src={(result as any).favicon || (result.url ? `https://www.google.com/s2/favicons?sz=32&domain_url=${encodeURIComponent(result.url)}` : "")}
                                alt="favicon"
                                className="h-4 w-4 rounded-sm shrink-0"
                                loading="lazy"
                              />
                            )}
                            <h5 className="text-xs font-semibold line-clamp-2">{result.title}</h5>
                          </div>
                        </div>
                        <div className="p-3 pt-0">
                          <p className="truncate text-xs text-gray-500">{result.url}</p>
                        </div>
                      </div>
                    </a>
                  </HoverCardTrigger>
                  <HoverCardContent className="w-80 max-h-56 overflow-y-auto">
                    <p className="text-sm">{result.content}</p>
                  </HoverCardContent>
                </HoverCard>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}