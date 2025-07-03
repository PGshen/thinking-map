import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';
import type { User } from '../types/user';

interface GlobalStore {
  user: User | null;
  mapId: string | null;
  loading: boolean;
  error: string | null;
  setUser: (user: User | null) => void;
  setMapId: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useGlobalStore = create<GlobalStore>()(
  devtools(
    persist(
      (set) => ({
        user: null,
        mapId: null,
        loading: false,
        error: null,
        setUser: (user) => set({ user }),
        setMapId: (id) => set({ mapId: id }),
        setLoading: (loading) => set({ loading }),
        setError: (error) => set({ error }),
      }),
      { name: 'global-store' }
    )
  )
); 