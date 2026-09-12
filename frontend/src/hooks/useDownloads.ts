import { useCallback, useEffect, useState } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { StartDownload, CancelDownload } from "../../wailsjs/go/main/App";
import { ytdlp } from "../../wailsjs/go/models";
import type { DownloadItem, JobProgress } from "../types/download";

export function useDownloads() {
    const [items, setItems] = useState<DownloadItem[]>([]);

    useEffect(() => {
        return EventsOn("download:progress", (state: JobProgress) => {
            setItems((prev) => {
                const idx = prev.findIndex((it) => it.JobID === state.JobID);
                if (idx === -1) {
                    // Shouldn't normally happen since startDownload registers
                    // the item locally before any event can arrive, but stay
                    // robust in case an event beats that registration.
                    return [...prev, { ...state, url: "", title: "" }];
                }
                const next = [...prev];
                next[idx] = { ...next[idx], ...state };
                return next;
            });
        });
    }, []);

    const startDownload = useCallback(
        async (url: string, title: string, options: ytdlp.DownloadOptions) => {
            const jobId = await StartDownload(options);
            setItems((prev) => [
                ...prev,
                {
                    JobID: jobId,
                    Status: "queued",
                    DownloadedBytes: 0,
                    TotalBytes: 0,
                    Percent: 0,
                    Speed: 0,
                    ETA: 0,
                    Message: "",
                    url,
                    title,
                },
            ]);
            return jobId;
        },
        []
    );

    const cancelDownload = useCallback((jobId: string) => {
        CancelDownload(jobId);
    }, []);

    return { items, startDownload, cancelDownload };
}
