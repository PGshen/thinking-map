import { EdgeProps, getBezierPath } from "reactflow";

// 自定义依赖边组件
export const DependencyEdge: React.FC<EdgeProps> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
}) => {
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <>
      <path
        id={id}
        style={{
          strokeDasharray: '5,5',
          stroke: '#8b5cf6',
          strokeWidth: 2,
          fill: 'none',
          ...style,
        }}
        className="react-flow__edge-path animated"
        d={edgePath}
        markerEnd={markerEnd}
      />
    </>
  );
};