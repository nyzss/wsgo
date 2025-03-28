import { useState, ChangeEvent } from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";

interface RoomManagerProps {
    isConnected: boolean;
    onJoinRoom: (roomId: string) => void;
    onLeaveRoom: (roomId: string) => void;
    onMessage: (message: string) => void;
}

export const RoomManager = ({
    isConnected,
    onJoinRoom,
    onLeaveRoom,
    onMessage,
}: RoomManagerProps) => {
    const [roomId, setRoomId] = useState("");
    const [activeRooms, setActiveRooms] = useState<string[]>([]);

    const handleJoinRoom = () => {
        if (roomId.trim()) {
            onJoinRoom(roomId);
            setActiveRooms([...activeRooms, roomId]);
            onMessage(`Joined room: ${roomId}`);
            setRoomId("");
        }
    };

    const handleLeaveRoom = (room: string) => {
        onLeaveRoom(room);
        setActiveRooms(activeRooms.filter((r) => r !== room));
        onMessage(`Left room: ${room}`);
    };

    return (
        <div className="custom-card rounded-lg border h-full flex flex-col">
            <div className="p-3 border-b border-accent/10">
                <h2 className="text-lg font-semibold text-white">
                    Room Management
                </h2>
            </div>

            <div className="p-3 border-b border-accent/10">
                <div className="flex items-center space-x-4">
                    <Input
                        type="text"
                        value={roomId}
                        onChange={(e: ChangeEvent<HTMLInputElement>) =>
                            setRoomId(e.target.value)
                        }
                        placeholder="Enter room ID"
                        disabled={!isConnected}
                        className="font-mono text-sm bg-[#252D3D] border-accent/10"
                    />
                    <Button
                        onClick={handleJoinRoom}
                        disabled={!isConnected || !roomId.trim()}
                        variant="default"
                        className="bg-[#374151] hover:bg-[#4B5563] shrink-0"
                    >
                        Join Room
                    </Button>
                </div>
            </div>

            <div className="px-3 py-2 border-b border-accent/10">
                <h3 className="text-sm font-medium text-muted-foreground">
                    Active Rooms
                </h3>
            </div>

            <div className="p-3 overflow-auto flex-1">
                {activeRooms.length > 0 ? (
                    <div className="space-y-2">
                        {activeRooms.map((room) => (
                            <div
                                key={room}
                                className="flex items-center justify-between p-3 rounded-md bg-[#252D3D] border border-accent/10"
                            >
                                <span className="font-mono text-sm break-all mr-2 flex-1">
                                    {room}
                                </span>
                                <Button
                                    onClick={() => handleLeaveRoom(room)}
                                    variant="destructive"
                                    size="sm"
                                    className="shrink-0"
                                >
                                    Leave
                                </Button>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                        No active rooms
                    </div>
                )}
            </div>
        </div>
    );
};
