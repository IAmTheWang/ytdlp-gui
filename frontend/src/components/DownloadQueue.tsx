import type { DownloadItem } from "../types/download";
import DownloadQueueItem from "./DownloadQueueItem";
import { useI18n } from "../i18n";

interface DownloadQueueProps {
    items: DownloadItem[];
    onCancel: (jobId: string) => void;
}

function DownloadQueue({ items, onCancel }: DownloadQueueProps) {
    const { t } = useI18n();
    if (items.length === 0) return null;

    return (
        <div className="download-queue">
            <h2>{t("queue.title")}</h2>
            {items.map((item) => (
                <DownloadQueueItem key={item.JobID} item={item} onCancel={onCancel} />
            ))}
        </div>
    );
}

export default DownloadQueue;
