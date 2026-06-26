import { create } from "zustand";
import type {
  LeaderboardEntry,
  LeaderboardPeriod,
  Submission,
  SubmissionStatus,
  TestResult,
  WsMessage,
} from "../types";
import { WS_BASE_URL } from "../services/api";
import { createBatcher } from "../hooks/useDebounce";

// --- WebSocket Configuration ---
const WS_CONFIG = {
  HEARTBEAT_INTERVAL: 30000, // 30s
  HEARTBEAT_TIMEOUT: 5000, // 5s
  MAX_RECONNECT_ATTEMPTS: 10,
  INITIAL_RECONNECT_DELAY: 1000, // 1s
  MAX_RECONNECT_DELAY: 30000, // 30s
  MESSAGE_BATCH_INTERVAL: 100, // 100ms batching
};

interface SubmissionState {
  currentSubmission: Submission | null;
  submissionHistory: Submission[];
  isSubmitting: boolean;
  wsConnection: WebSocket | null;
  isConnected: boolean;
  liveStatus: SubmissionStatus | null;
  progress: number;
  completedTests: number;
  totalTests: number;
  leaderboard: LeaderboardEntry[];
  leaderboardPeriod: LeaderboardPeriod;
  reconnectAttempts: number;

  // Actions
  setCurrentSubmission: (submission: Submission | null) => void;
  updateSubmissionStatus: (status: SubmissionStatus) => void;
  addTestResult: (result: TestResult) => void;
  setSubmitting: (val: boolean) => void;
  setProgress: (completed: number, total: number) => void;
  setLeaderboard: (entries: LeaderboardEntry[]) => void;
  setLeaderboardPeriod: (period: LeaderboardPeriod) => void;
  addToHistory: (submission: Submission) => void;
  setSubmissionHistory: (history: Submission[]) => void;

  // WebSocket
  connectWebSocket: (token: string) => void;
  disconnectWebSocket: () => void;
  subscribeToSubmission: (submissionId: string) => void;
  manualReconnect: () => void;
}

// Message batcher for incoming WebSocket messages
let messageBatcher: ReturnType<typeof createBatcher<WsMessage>> | null = null;

function createMessageBatcher(handler: (messages: WsMessage[]) => void) {
  return createBatcher<WsMessage>(handler, WS_CONFIG.MESSAGE_BATCH_INTERVAL);
}

export const useSubmissionStore = create<SubmissionState>((set, get) => {
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  let heartbeatTimeoutTimer: ReturnType<typeof setTimeout> | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let reconnectAttempts = 0;
  let intentionalClose = false;
  let currentToken = "";

  const clearTimers = () => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer);
      heartbeatTimer = null;
    }
    if (heartbeatTimeoutTimer) {
      clearTimeout(heartbeatTimeoutTimer);
      heartbeatTimeoutTimer = null;
    }
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  };

  const getReconnectDelay = (attempt: number) => {
    // Exponential backoff with jitter
    const base = WS_CONFIG.INITIAL_RECONNECT_DELAY * Math.pow(2, attempt);
    const jitter = Math.random() * 1000;
    return Math.min(base + jitter, WS_CONFIG.MAX_RECONNECT_DELAY);
  };

  const startHeartbeat = (ws: WebSocket) => {
    heartbeatTimer = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping", timestamp: new Date().toISOString() }));

        // Set heartbeat timeout - if no pong received, close connection
        heartbeatTimeoutTimer = setTimeout(() => {
          console.warn("[WS] Heartbeat timeout, closing connection");
          ws.close(1000, "Heartbeat timeout");
        }, WS_CONFIG.HEARTBEAT_TIMEOUT);
      }
    }, WS_CONFIG.HEARTBEAT_INTERVAL);
  };

  const attemptReconnect = (token: string) => {
    const state = get();
    if (state.wsConnection?.readyState === WebSocket.OPEN) return;

    if (reconnectAttempts >= WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
      console.warn("[WS] Max reconnect attempts reached");
      return;
    }

    const delay = getReconnectDelay(reconnectAttempts);
    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${reconnectAttempts + 1})`);

    reconnectTimer = setTimeout(() => {
      reconnectAttempts += 1;
      set({ reconnectAttempts });

      // Close existing before reconnecting
      const existing = get().wsConnection;
      if (existing) {
        existing.onopen = null;
        existing.onmessage = null;
        existing.onclose = null;
        existing.onerror = null;
        existing.close();
      }

      const ws = createWebSocket(token);
      set({ wsConnection: ws });
    }, delay);
  };

  const createWebSocket = (token: string): WebSocket => {
    const ws = new WebSocket(`${WS_BASE_URL}?token=${token}`);

    ws.onopen = () => {
      console.log("[WS] Connected");
      reconnectAttempts = 0;
      set({ isConnected: true, reconnectAttempts: 0 });
      clearTimers();

      // Start heartbeat
      startHeartbeat(ws);
    };

    ws.onmessage = (event: MessageEvent) => {
      // Respond to pong/heartbeat
      try {
        const raw = JSON.parse(event.data);
        if (raw.type === "pong") {
          if (heartbeatTimeoutTimer) {
            clearTimeout(heartbeatTimeoutTimer);
            heartbeatTimeoutTimer = null;
          }
          return;
        }
      } catch {
        // Not JSON, ignore
      }

      // Use message batcher for non-heartbeat messages
      if (messageBatcher) {
        try {
          const message: WsMessage = JSON.parse(event.data);
          messageBatcher.add(message);
        } catch {
          // ignore malformed messages
        }
      }
    };

    ws.onclose = (event: CloseEvent) => {
      console.log(`[WS] Closed: ${event.code} ${event.reason}`);
      clearTimers();

      if (!intentionalClose) {
        // Auto-reconnect with exponential backoff
        attemptReconnect(currentToken);
      } else {
        set({ isConnected: false, wsConnection: null });
      }
    };

    ws.onerror = () => {
      console.warn("[WS] Error occurred");
      // onclose will fire after onerror, so reconnect is handled there
    };

    return ws;
  };

  // Initialize message batcher with state handlers
  const processBatchedMessages = (messages: WsMessage[]) => {
    const state = get();
    for (const message of messages) {
      try {
        switch (message.type) {
          case "submission_update": {
            const payload = message.payload as {
              status: SubmissionStatus;
              progress: number;
              completed_tests: number;
              total_tests: number;
            };
            state.updateSubmissionStatus(payload.status);
            state.setProgress(payload.completed_tests, payload.total_tests);
            break;
          }
          case "test_result": {
            const payload = message.payload as TestResult;
            state.addTestResult(payload);
            break;
          }
          case "leaderboard_update": {
            const payload = message.payload as { entries: LeaderboardEntry[] };
            state.setLeaderboard(payload.entries);
            break;
          }
          case "connected":
            break;
          case "error":
            break;
        }
      } catch {
        // ignore individual message errors
      }
    }
  };

  // Initialize batcher once
  messageBatcher = createMessageBatcher(processBatchedMessages);

  return {
    currentSubmission: null,
    submissionHistory: [],
    isSubmitting: false,
    wsConnection: null,
    isConnected: false,
    liveStatus: null,
    progress: 0,
    completedTests: 0,
    totalTests: 0,
    leaderboard: [],
    leaderboardPeriod: "all-time",
    reconnectAttempts: 0,

    setCurrentSubmission: (submission) => set({ currentSubmission: submission }),
    updateSubmissionStatus: (status) => set({ liveStatus: status }),
    addTestResult: (result) => {
      const { currentSubmission } = get();
      if (currentSubmission) {
        const updatedResults = [...currentSubmission.test_results, result];
        set({
          currentSubmission: { ...currentSubmission, test_results: updatedResults },
          completedTests: updatedResults.length,
        });
      }
    },
    setSubmitting: (val) => set({ isSubmitting: val }),
    setProgress: (completed, total) =>
      set({
        completedTests: completed,
        totalTests: total,
        progress: total > 0 ? (completed / total) * 100 : 0,
      }),
    setLeaderboard: (entries) => set({ leaderboard: entries }),
    setLeaderboardPeriod: (period) => set({ leaderboardPeriod: period }),
    addToHistory: (submission) =>
      set((state) => ({
        submissionHistory: [submission, ...state.submissionHistory].slice(0, 50),
      })),
    setSubmissionHistory: (history) => set({ submissionHistory: history }),

    connectWebSocket: (token: string) => {
      currentToken = token;
      intentionalClose = false;
      const { wsConnection } = get();
      if (wsConnection?.readyState === WebSocket.OPEN) return;

      const ws = createWebSocket(token);
      set({ wsConnection: ws });
    },

    disconnectWebSocket: () => {
      intentionalClose = true;
      clearTimers();
      const { wsConnection } = get();
      if (wsConnection) {
        wsConnection.close(1000, "Intentional disconnect");
        set({ wsConnection: null, isConnected: false });
      }
    },

    subscribeToSubmission: (submissionId: string) => {
      const { wsConnection } = get();
      if (wsConnection?.readyState === WebSocket.OPEN) {
        wsConnection.send(
          JSON.stringify({ type: "subscribe", submission_id: submissionId })
        );
      }
    },

    manualReconnect: () => {
      intentionalClose = false;
      const { wsConnection } = get();
      if (wsConnection) {
        wsConnection.onopen = null;
        wsConnection.onmessage = null;
        wsConnection.onclose = null;
        wsConnection.onerror = null;
        wsConnection.close();
      }
      if (currentToken) {
        attemptReconnect(currentToken);
      }
    },
  };
});
