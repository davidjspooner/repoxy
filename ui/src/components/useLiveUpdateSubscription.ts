import { useEffect } from 'react';

type Callback = () => void;

export interface LiveUpdateOptions {
  onConnect?: Callback;
  onDisconnect?: Callback;
  onData?: (payload: unknown) => void;
}

/**
 * Placeholder hook that simulates a live update subscription.
 * Replace with real WebSocket/SSE logic when backend endpoints are ready.
 */
export function useLiveUpdateSubscription({ onConnect, onDisconnect, onData }: LiveUpdateOptions = {}) {
  useEffect(() => {
    const timeout = setTimeout(() => {
      onConnect?.();
      onData?.({ type: 'sample', timestamp: Date.now() });
    }, 1000);

    return () => {
      clearTimeout(timeout);
      onDisconnect?.();
    };
  }, [onConnect, onDisconnect, onData]);
}
