import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';
import type { Node, Edge } from '../types/node';

interface MapStore {
  nodes: Node[];
  edges: Edge[];
  selectedNodeId: string | null;
  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  selectNode: (id: string | null) => void;
}

export const useMapStore = create<MapStore>()(
  devtools(
    persist(
      (set) => ({
        nodes: [],
        edges: [],
        selectedNodeId: null,
        setNodes: (nodes) => set({ nodes }),
        setEdges: (edges) => set({ edges }),
        selectNode: (id) => set({ selectedNodeId: id }),
      }),
      { name: 'map-store' }
    )
  )
); 