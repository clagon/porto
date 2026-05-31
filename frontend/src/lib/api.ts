import type { HealthResponse, PortMapping, Settings, StatusResponse } from './types';

function browserToken(): string {
  if (typeof document === 'undefined') {
    return '';
  }
  return document.querySelector<HTMLMetaElement>('meta[name="porto-browser-token"]')?.content ?? '';
}

function requestHeaders(init?: RequestInit): Headers {
  const headers = new Headers(init?.headers);
  headers.set('Content-Type', 'application/json');
  const token = browserToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  return headers;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: requestHeaders(init),
  });
  if (!res.ok) {
    let errMsg = `${path} failed: ${res.status}`;
    try {
      const rawText = await res.text();
      if (rawText) {
        try {
          const errObj = JSON.parse(rawText);
          if (errObj && typeof errObj === 'object' && 'message' in errObj) {
            errMsg = String(errObj.message);
          } else {
            errMsg = rawText;
          }
        } catch {
          errMsg = rawText;
        }
      }
    } catch {
      errMsg = `${path} failed: ${res.status}`;
    }
    throw new Error(errMsg);
  }
  return (await res.json()) as T;
}

export const api = {
  health: () => request<HealthResponse>('/api/health'),
  status: () => request<StatusResponse>('/api/status'),
  discover: () => request<StatusResponse>('/api/discover', { method: 'POST' }),
  openPort: (mapping: PortMapping) => request<StatusResponse>('/api/ports/open', {
    method: 'POST',
    body: JSON.stringify(mapping),
  }),
  closePort: (mapping: PortMapping) => request<StatusResponse>('/api/ports/close', {
    method: 'POST',
    body: JSON.stringify(mapping),
  }),
  getSettings: () => request<Settings>('/api/settings'),
  saveSettings: (settings: Settings) => request<{ ok: boolean }>('/api/settings', {
    method: 'POST',
    body: JSON.stringify(settings),
  }),
};
