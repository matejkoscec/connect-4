import { createFileRoute, Link } from "@tanstack/react-router";
import { useCallback, useState, useRef, FormEvent } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useGameSocket, GameSocketHandlers } from "@/hooks/useGameSocket";
import type {
  FoundGamePayload,
  GameOverPayload,
  ChatMessagePayload,
} from "@/api/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/games")({
  component: GamePageComponent,
});

const ROWS = 6;
const COLS = 7;
const PLAYER_COLORS = ["", "bg-red-500", "bg-yellow-500"];

type GameStatus = "idle" | "waiting" | "playing" | "over";

function GamePageComponent() {
  const { isAuthenticated, user, isLoading } = useAuth();
  const [status, setStatus] = useState<GameStatus>("idle");
  const [gameInfo, setGameInfo] = useState<FoundGamePayload | null>(null);
  const [board, setBoard] = useState<number[][]>(() =>
    Array.from({ length: ROWS }, () => Array(COLS).fill(0))
  );
  const [activePlayer, setActivePlayer] = useState<number>(1);
  const [winner, setWinner] = useState<GameOverPayload | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessagePayload[]>([]);
  const [chatInput, setChatInput] = useState("");
  const chatScrollAreaRef = useRef<HTMLDivElement>(null);

  const handlers = useCallback(
    (): GameSocketHandlers => ({
      onWaitingForGame: () => setStatus("waiting"),
      onFoundGame: (payload) => {
        setGameInfo(payload);
        setStatus("playing");
        setWinner(null);
        setBoard(() => {
          const b = Array.from({ length: ROWS }, () => Array(COLS).fill(0));
          for (let i = 0; i < ROWS; i++) {
            for (let j = 0; j < COLS; j++) {
              b[i][j] = payload.state[i][j];
            }
          }
          return b;
        });
        setActivePlayer(
          payload.lastPlayed !== 0 ? (payload.lastPlayed === 1 ? 2 : 1) : 1
        );
        setChatMessages(payload.messages);
      },

      onPlayedMove: (payload) => {
        setBoard((prevBoard) => {
          const newBoard = prevBoard.map((row) => [...row]);

          newBoard[payload.row][payload.column] = payload.color;
          return newBoard;
        });

        setActivePlayer((prev) => (prev === 1 ? 2 : 1));
      },

      onGameOver: (payload) => {
        setWinner(payload);
        setStatus("over");
      },
      onChatMessage: (payload) => {
        setChatMessages((prev) => [...prev, payload]);
      },
    }),
    []
  );

  const { sendWaitingForGame, sendPlayMove, sendChatMessage, connected } =
    useGameSocket(handlers());

  const handleFindGame = () => {
    sendWaitingForGame();
    setStatus("waiting");
  };

  const handlePlayMove = (colIndex: number) => {
    if (status !== "playing" || !gameInfo || gameInfo.color !== activePlayer)
      return;
    sendPlayMove(colIndex);
  };

  const handleSendChat = (e: FormEvent) => {
    e.preventDefault();
    if (chatInput.trim() && user) {
      sendChatMessage(user.username, chatInput.trim());
      setChatInput("");
    }
  };

  if (isLoading) return <p className="p-4">Loading authentication...</p>;
  if (!isAuthenticated || !user) {
    return (
      <div className="p-4 text-center">
        <p>You must be logged in to play.</p>
        <Button asChild className="mt-2">
          <Link to="/login">Go to Login</Link>
        </Button>
      </div>
    );
  }

  const renderStatusMessage = () => {
    switch (status) {
      case "idle":
        return 'Click "Find Game" to start.';
      case "waiting":
        return "Searching for an opponent...";
      case "playing":
        if (winner) {
          return winner.winner === gameInfo?.color
            ? "You won! 🎉"
            : "You lost. 😥";
        }
        if (!gameInfo) return "Setting up game...";
        return gameInfo.color === activePlayer
          ? "It's your turn!"
          : "Waiting for opponent's move...";
      case "over":
        if (winner?.winner === 0) return "It's a draw!";
        return winner?.winner === gameInfo?.color
          ? "You won! 🎉"
          : "You lost. 😥";
    }
  };

  return (
    <div className="container mx-auto p-4 grid grid-cols-1 lg:grid-cols-3 gap-6 h-[calc(100vh-80px)]">
      {/* Game Board and Actions */}
      <div className="lg:col-span-2 flex flex-col items-center justify-center">
        {status === "idle" || status === "waiting" ? (
          <div className="text-center">
            <h2 className="text-2xl font-bold mb-4">{renderStatusMessage()}</h2>
            <Button
              onClick={handleFindGame}
              size="lg"
              disabled={status === "waiting"}
            >
              {status === "waiting" ? "Searching..." : "Find Game"}
            </Button>
          </div>
        ) : (
          <div className="w-full max-w-3xl aspect-[7/6] bg-blue-700 p-4 rounded-lg shadow-2xl grid grid-cols-7 gap-2">
            {board.flat().map((cell, index) => (
              <div
                key={index}
                className="relative rounded-full bg-blue-900 cursor-pointer group"
                onClick={() => handlePlayMove(index % COLS)}
              >
                <div
                  className={cn(
                    "absolute inset-0 rounded-full transition-transform duration-300 ease-out",
                    PLAYER_COLORS[cell],
                    status === "playing" &&
                      gameInfo?.color === activePlayer &&
                      cell === 0 &&
                      "group-hover:bg-gray-400/50"
                  )}
                />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Info and Chat Panel */}
      <div className="flex flex-col gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Game Info</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              <strong>Status:</strong> {renderStatusMessage()}
            </p>
            {gameInfo && (
              <p>
                <strong>Your Color:</strong>
                <span
                  className={cn(
                    "font-bold",
                    gameInfo.color === 1 ? "text-red-500" : "text-yellow-500"
                  )}
                >
                  {gameInfo.color === 1 ? " Red" : " Yellow"}
                </span>
              </p>
            )}
            {status === "over" && (
              <Button onClick={handleFindGame} className="w-full mt-4">
                Play Again
              </Button>
            )}
          </CardContent>
        </Card>
        <Card className="flex-1 flex flex-col">
          <CardHeader>
            <CardTitle>Chat</CardTitle>
          </CardHeader>
          <CardContent className="flex-1 flex flex-col gap-2 p-0">
            <ScrollArea className="flex-1 p-6" ref={chatScrollAreaRef}>
              {chatMessages.map((msg, i) => (
                <div key={i} className="mb-2">
                  <span className="font-bold">{msg.from}: </span>
                  <span>{msg.text}</span>
                </div>
              ))}
            </ScrollArea>
            <form onSubmit={handleSendChat} className="p-4 border-t flex gap-2">
              <Input
                placeholder="Type a message..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                disabled={status !== "playing" && status !== "over"}
              />
              <Button type="submit" disabled={!chatInput.trim() || !connected}>
                Send
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
