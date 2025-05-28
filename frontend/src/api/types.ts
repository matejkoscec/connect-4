export const Color = ["none", "Red", "Yellow"];

export const MESSAGE_TYPES = {
  WAITING_FOR_GAME: "waitingForGame",
  FOUND_GAME: "foundGame",
  CHAT_MESSAGE: "chatMessage",
  PLAY_MOVE: "playMove",
  PLAYED_MOVE: "playedMove",
  GAME_OVER: "gameOver",
} as const;

export type MessageType = (typeof MESSAGE_TYPES)[keyof typeof MESSAGE_TYPES];

export interface WaitingForGamePayload {}

export interface FoundGamePayload {
  lobbyId: string;
  color: number;
  state: number[][];
  lastPlayed: number;
  messages: ChatMessagePayload[];
}

export interface ChatMessagePayload {
  from: string;
  text: string;
}

export interface PlayMovePayload {
  column: number;
}

export interface PlayedMovePayload {
  color: number;
  row: number;
  column: number;
}

export interface GameOverPayload {
  winner: number;
}

export type Payload =
  | WaitingForGamePayload
  | FoundGamePayload
  | ChatMessagePayload
  | PlayMovePayload
  | PlayedMovePayload
  | GameOverPayload;

export interface Message<T extends Payload = Payload> {
  version: string;
  type: MessageType;
  payload: T;
}

export type ErrorPayload<T extends Payload = Payload> = {
  code: number;
  err: string;
  problematicMsg: Message<T>;
};

export interface WaitingForGameMessage extends Message<WaitingForGamePayload> {
  type: typeof MESSAGE_TYPES.WAITING_FOR_GAME;
}

export interface FoundGameMessage extends Message<FoundGamePayload> {
  type: typeof MESSAGE_TYPES.FOUND_GAME;
}

export interface ChatMessage extends Message<ChatMessagePayload> {
  type: typeof MESSAGE_TYPES.CHAT_MESSAGE;
}

export interface PlayMoveMessage extends Message<PlayMovePayload> {
  type: typeof MESSAGE_TYPES.PLAY_MOVE;
}

export interface PlayedMoveMessage extends Message<PlayedMovePayload> {
  type: typeof MESSAGE_TYPES.PLAYED_MOVE;
}

export interface GameOverMessage extends Message<GameOverPayload> {
  type: typeof MESSAGE_TYPES.GAME_OVER;
}

export type WebsocketMessage =
  | WaitingForGameMessage
  | FoundGameMessage
  | ChatMessage
  | PlayMoveMessage
  | PlayedMoveMessage
  | GameOverMessage;

export const isWaitingForGameMessage = (
  msg: WebsocketMessage
): msg is WaitingForGameMessage => msg.type === MESSAGE_TYPES.WAITING_FOR_GAME;

export const isFoundGameMessage = (
  msg: WebsocketMessage
): msg is FoundGameMessage => msg.type === MESSAGE_TYPES.FOUND_GAME;

export const isChatMessage = (msg: WebsocketMessage): msg is ChatMessage =>
  msg.type === MESSAGE_TYPES.CHAT_MESSAGE;

export const isPlayMoveMessage = (
  msg: WebsocketMessage
): msg is PlayMoveMessage => msg.type === MESSAGE_TYPES.PLAY_MOVE;

export const isPlayedMoveMessage = (
  msg: WebsocketMessage
): msg is PlayedMoveMessage => msg.type === MESSAGE_TYPES.PLAYED_MOVE;

export const isGameOverMessage = (
  msg: WebsocketMessage
): msg is GameOverMessage => msg.type === MESSAGE_TYPES.GAME_OVER;

export const createMessage = {
  waitingForGame: (): WaitingForGameMessage => ({
    version: "v1",
    type: MESSAGE_TYPES.WAITING_FOR_GAME,
    payload: {},
  }),

  chatMessage: (from: string, text: string): ChatMessage => ({
    version: "v1",
    type: MESSAGE_TYPES.CHAT_MESSAGE,
    payload: { from, text },
  }),

  playMove: (column: number): PlayMoveMessage => ({
    version: "v1",
    type: MESSAGE_TYPES.PLAY_MOVE,
    payload: { column },
  }),

  playedMove: (
    color: number,
    row: number,
    column: number
  ): PlayedMoveMessage => ({
    version: "v1",
    type: MESSAGE_TYPES.PLAYED_MOVE,
    payload: { color, row, column },
  }),

  gameOver: (winner: number): GameOverMessage => ({
    version: "v1",
    type: MESSAGE_TYPES.GAME_OVER,
    payload: { winner },
  }),
};
