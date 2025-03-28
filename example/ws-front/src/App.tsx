import { useState } from "react";
import { WebSocketManager } from "./components/WebSocketManager";
import { LogViewer } from "./components/LogViewer";
import { RoomManager } from "./components/RoomManager";

function App() {
    const [logs, setLogs] = useState<string[]>([]);
    const [isConnected, setIsConnected] = useState(false);

    const handleMessage = (message: string) => {
        setLogs((prev) => [...prev, message]);
    };

    const handleConnectionChange = (connected: boolean) => {
        setIsConnected(connected);
    };

    const handleJoinRoom = (roomId: string) => {
        // TODO: add join logic
        handleMessage(`Attempting to join room: ${roomId}`);
    };

    const handleLeaveRoom = (roomId: string) => {
        // TODO: add leave logic
        handleMessage(`Attempting to leave room: ${roomId}`);
    };

    const handleClearLogs = () => {
        setLogs([]);
    };

    return (
        <div className="h-screen flex flex-col overflow-hidden bg-[#121620] text-foreground">
            <div className="container mx-auto py-3 px-4 flex-1 flex flex-col overflow-hidden">
                <div className="flex items-center justify-between mb-3">
                    <h1 className="text-2xl font-bold text-white">
                        WebSocket Testing Interface
                    </h1>
                </div>

                <WebSocketManager
                    onMessage={handleMessage}
                    onConnectionChange={handleConnectionChange}
                />

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 flex-1 h-full min-h-0">
                    <RoomManager
                        isConnected={isConnected}
                        onJoinRoom={handleJoinRoom}
                        onLeaveRoom={handleLeaveRoom}
                        onMessage={handleMessage}
                    />

                    <LogViewer logs={logs} onClear={handleClearLogs} />
                </div>
            </div>
        </div>
    );
}

export default App;
