import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Normalize various tag shapes into a string array
export function normalizeTagsArray(raw: any): string[] {
  if (!raw) return [];
  if (Array.isArray(raw)) {
    return raw
      .map((t: any) => {
        if (!t && t !== 0) return '';
        if (typeof t === 'string') return t;
        if (typeof t === 'number') return String(t);
        if (typeof t === 'object') {
          return t.name || t.Name || t.title || t.label || (t.id ? String(t.id) : '') || '';
        }
        return String(t);
      })
      .map((s: string) => (s ? s.trim() : ''))
      .filter(Boolean);
  }
  if (typeof raw === 'string') {
    return raw.split(',').map((s: string) => s.trim()).filter(Boolean);
  }
  return [];
}
