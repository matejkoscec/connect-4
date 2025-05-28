import { useEffect, useRef, useCallback, useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import paths from "@/api/paths";
import {
  WebsocketMessage,
  isFoundGameMessage,
  isPlayedMoveMessage,
  isGameOverMessage,
  isChatMessage,
  createMessage,
  FoundGamePayload,
  PlayedMovePayload,
  GameOverPayload,
  ChatMessagePayload,
  WaitingForGamePayload,
  isWaitingForGameMessage,
} from "@/api/types";

export type GameSocketHandlers = {
  onFoundGame: (payload: FoundGamePayload) => void;
  onPlayedMove: (payload: PlayedMovePayload) => void;
  onGameOver: (payload: GameOverPayload) => void;
  onChatMessage: (payload: ChatMessagePayload) => void;
  onWaitingForGame?: (payload: WaitingForGamePayload) => void;
  onOpen?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (event: Event) => void;
};

export const useGameSocket = (handlers: GameSocketHandlers) => {
  const { token, isAuthenticated } = useAuth();
  const socketRef = useRef<WebSocket | null>(null);
  const handlersRef = useRef(handlers);
  const [look, setLook] = useState(false);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    handlersRef.current = handlers;
  }, [handlers]);

  useEffect(() => {
    if (!isAuthenticated || !token) return;
    if (!look) return;

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const host = "localhost:8080";
    const wsUrl = `${protocol}://${host}${paths.base}${paths.games.play}?token=${token}`;

    const ws = new WebSocket(wsUrl);
    socketRef.current = ws;

    ws.onopen = (event) => {
      console.log("✅ WebSocket connection established.");
      setConnected(true);
      handlersRef.current.onOpen?.(event);
    };

    ws.onmessage = (event) => {
      try {
        const message: WebsocketMessage = JSON.parse(event.data as string);
        console.log(message);

        if (isWaitingForGameMessage(message)) {
          handlersRef.current.onWaitingForGame?.(message.payload);
        } else if (isFoundGameMessage(message)) {
          handlersRef.current.onFoundGame(message.payload);
        } else if (isPlayedMoveMessage(message)) {
          handlersRef.current.onPlayedMove(message.payload);
        } else if (isGameOverMessage(message)) {
          handlersRef.current.onGameOver(message.payload);
        } else if (isChatMessage(message)) {
          handlersRef.current.onChatMessage(message.payload);
        }
      } catch (error) {
        console.error("❌ Error parsing WebSocket message:", error);
      }
    };

    ws.onerror = (event) => {
      console.error("❌ WebSocket error:", event);
      handlersRef.current.onError?.(event);
    };

    ws.onclose = (event) => {
      console.log("⚪ WebSocket connection closed.", event.reason);
      socketRef.current = null;
      handlersRef.current.onClose?.(event);
      setLook(false);
      setConnected(false);
    };

    return () => {
      socketRef.current?.close(1000, "Component unmounting");
    };
  }, [token, isAuthenticated, look]);

  /**
   * Sends a message through the WebSocket if the connection is open.
   * @param message - The message object to send.
   */
  const sendMessage = useCallback((message: WebsocketMessage) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(message));
    } else {
      console.error("WebSocket is not open. Cannot send message.");
    }
  }, []);

  const sendWaitingForGame = useCallback(() => {
    setLook(true);
  }, [sendMessage]);

  const sendPlayMove = useCallback(
    (column: number) => {
      sendMessage(createMessage.playMove(column));
    },
    [sendMessage]
  );

  const sendChatMessage = useCallback(
    (from: string, text: string) => {
      sendMessage(createMessage.chatMessage(from, text));
    },
    [sendMessage]
  );

  return {
    readyState: socketRef.current?.readyState ?? WebSocket.CLOSED,
    connected,
    sendWaitingForGame,
    sendPlayMove,
    sendChatMessage,
  };
};
