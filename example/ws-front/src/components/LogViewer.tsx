import { Button } from "./ui/button";

interface LogViewerProps {
    logs: string[];
    onClear: () => void;
}

export const LogViewer = ({ logs, onClear }: LogViewerProps) => {
    return (
        <div className="custom-card rounded-lg border h-full flex flex-col">
            <div className="flex items-center justify-between p-3 border-b border-accent/10">
                <h2 className="text-lg font-semibold text-white">
                    WebSocket Logs
                </h2>
                <Button variant="destructive" size="sm" onClick={onClear}>
                    Clear Logs
                </Button>
            </div>
            <div className="p-3 overflow-auto flex-1">
                <div className="space-y-2">
                    {logs.map((log, index) => (
                        <div
                            key={index}
                            className="text-sm font-mono py-1 px-2 rounded bg-[#252D3D]"
                        >
                            <span className="text-muted-foreground whitespace-nowrap">
                                [{new Date().toLocaleTimeString()}]
                            </span>{" "}
                            <span className="break-all">{log}</span>
                        </div>
                    ))}
                    {logs.length === 0 && (
                        <div className="text-center py-8 text-muted-foreground">
                            No logs yet. Connect to a WebSocket server to see
                            logs.
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
