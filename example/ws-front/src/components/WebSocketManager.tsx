import {
    useState,
    useEffect,
    useCallback,
    ChangeEvent,
    KeyboardEvent,
} from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";

interface WebSocketManagerProps {
    onMessage: (message: string) => void;
    onConnectionChange: (connected: boolean) => void;
}

export const WebSocketManager = ({
    onMessage,
    onConnectionChange,
}: WebSocketManagerProps) => {
    const [ws, setWs] = useState<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [url, setUrl] = useState("ws://localhost:8080");
    const [message, setMessage] = useState("");

    const connect = useCallback(() => {
        try {
            const websocket = new WebSocket(url);

            websocket.onopen = () => {
                setIsConnected(true);
                onConnectionChange(true);
                onMessage("Connected to WebSocket server");
            };

            websocket.onclose = () => {
                setIsConnected(false);
                onConnectionChange(false);
                onMessage("Disconnected from WebSocket server");
            };

            websocket.onerror = (error) => {
                onMessage(`WebSocket error: ${error}`);
            };

            websocket.onmessage = (event) => {
                onMessage(`Received: ${event.data}`);
            };

            setWs(websocket);
        } catch (error) {
            onMessage(`Connection error: ${error}`);
        }
    }, [url, onMessage, onConnectionChange]);

    const disconnect = useCallback(() => {
        if (ws) {
            ws.close();
            setWs(null);
        }
    }, [ws]);

    const sendMessage = useCallback(() => {
        if (ws && message.trim()) {
            ws.send(message);
            onMessage(`Sent: ${message}`);
            setMessage("");
        }
    }, [ws, message, onMessage]);

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    };

    useEffect(() => {
        return () => {
            if (ws) {
                ws.close();
            }
        };
    }, [ws]);

    return (
        <div className="custom-card rounded-lg border mb-4">
            <div className="p-3">
                <div className="flex flex-col sm:flex-row gap-3">
                    <div className="flex-1 flex items-center gap-3">
                        <div className="flex items-center space-x-2 min-w-[120px]">
                            <div
                                className={`w-2 h-2 rounded-full transition-colors ${
                                    isConnected
                                        ? "bg-green-500 shadow-[0_0_8px_0_rgba(34,197,94,0.4)]"
                                        : "bg-red-500 shadow-[0_0_8px_0_rgba(239,68,68,0.4)]"
                                }`}
                            />
                            <span className="text-sm text-muted-foreground">
                                {isConnected ? "Connected" : "Disconnected"}
                            </span>
                        </div>
                        <Input
                            type="text"
                            value={url}
                            onChange={(e: ChangeEvent<HTMLInputElement>) =>
                                setUrl(e.target.value)
                            }
                            placeholder="WebSocket URL"
                            className="font-mono text-sm bg-[#252D3D] border-accent/10"
                        />
                    </div>
                    <div className="flex space-x-3">
                        <Button
                            onClick={connect}
                            disabled={isConnected}
                            variant="default"
                            className="flex-1 bg-[#374151] hover:bg-[#4B5563]"
                        >
                            Connect
                        </Button>
                        <Button
                            onClick={disconnect}
                            disabled={!isConnected}
                            variant="destructive"
                            className="flex-1"
                        >
                            Disconnect
                        </Button>
                    </div>
                </div>

                {isConnected && (
                    <div className="mt-3">
                        <div className="flex items-center space-x-3">
                            <Input
                                type="text"
                                value={message}
                                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                                    setMessage(e.target.value)
                                }
                                onKeyDown={handleKeyDown}
                                placeholder="Enter message to send"
                                className="font-mono text-sm bg-[#252D3D] border-accent/10"
                            />
                            <Button
                                onClick={sendMessage}
                                disabled={!message.trim()}
                                className="bg-[#374151] hover:bg-[#4B5563]"
                            >
                                Send
                            </Button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};
