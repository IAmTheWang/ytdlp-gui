import { useEffect, useState } from "react";
import {
    GetSettings,
    SaveSettings,
    CheckBinaries,
    ChooseDownloadDir,
    ChooseFile,
} from "../../wailsjs/go/main/App";
import { config, binmanager } from "../../wailsjs/go/models";
import { LANGUAGES, isLanguage, useI18n } from "../i18n";

interface SettingsPanelProps {
    onClose: () => void;
}

function SettingsPanel({ onClose }: SettingsPanelProps) {
    const { t, setLang } = useI18n();
    const [settings, setSettings] = useState<config.Settings | null>(null);
    const [binStatus, setBinStatus] = useState<binmanager.BinaryStatus | null>(null);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<string | null>(null);

    useEffect(() => {
        GetSettings().then(setSettings);
        CheckBinaries().then(setBinStatus);
    }, []);

    function set<K extends keyof config.Settings>(key: K, value: config.Settings[K]) {
        setSettings((prev) => (prev ? ({ ...prev, [key]: value } as config.Settings) : prev));
    }

    async function handleSave() {
        if (!settings) return;
        setSaving(true);
        setMessage(null);
        try {
            await SaveSettings(settings);
            const saved = t("settings.saved");
            setMessage(saved);
            setBinStatus(await CheckBinaries());
            setTimeout(() => setMessage((m) => (m === saved ? null : m)), 2500);
        } catch (err) {
            setMessage(String(err));
        } finally {
            setSaving(false);
        }
    }

    async function pickDownloadDir() {
        const dir = await ChooseDownloadDir(t("settings.chooseDownloadDir"));
        if (dir) set("DownloadDir", dir);
    }

    async function pickFile(key: "YtDlpPath" | "FFmpegPath", title: string) {
        const path = await ChooseFile(title);
        if (path) set(key, path);
    }

    if (!settings) {
        return (
            <div className="settings-panel">
                <p>{t("settings.loading")}</p>
            </div>
        );
    }

    return (
        <div className="settings-panel">
            <div className="settings-header">
                <h2>{t("settings.title")}</h2>
                <button onClick={onClose}>{t("settings.back")}</button>
            </div>

            <label className="settings-field">
                <span>{t("settings.downloadDir")}</span>
                <div className="settings-path-row">
                    <input value={settings.DownloadDir} onChange={(e) => set("DownloadDir", e.target.value)} />
                    <button onClick={pickDownloadDir}>{t("settings.browse")}</button>
                </div>
            </label>

            <label className="settings-field">
                <span>{t("settings.outputTemplate")}</span>
                <input value={settings.OutputTemplate} onChange={(e) => set("OutputTemplate", e.target.value)} />
            </label>

            <label className="settings-field">
                <span>{t("settings.concurrency")}</span>
                <input
                    type="number"
                    min={1}
                    value={settings.Concurrency}
                    onChange={(e) => set("Concurrency", Number(e.target.value))}
                />
            </label>

            <label className="settings-field">
                <span>{t("settings.proxy")}</span>
                <input
                    placeholder="socks5://127.0.0.1:1080"
                    value={settings.Proxy}
                    onChange={(e) => set("Proxy", e.target.value)}
                />
            </label>

            <label className="settings-field">
                <span>{t("settings.ytdlpPath")}</span>
                <div className="settings-path-row">
                    <input value={settings.YtDlpPath} onChange={(e) => set("YtDlpPath", e.target.value)} />
                    <button onClick={() => pickFile("YtDlpPath", t("settings.chooseYtdlp"))}>
                        {t("settings.browse")}
                    </button>
                </div>
            </label>

            <label className="settings-field">
                <span>{t("settings.ffmpegPath")}</span>
                <div className="settings-path-row">
                    <input value={settings.FFmpegPath} onChange={(e) => set("FFmpegPath", e.target.value)} />
                    <button onClick={() => pickFile("FFmpegPath", t("settings.chooseFfmpeg"))}>
                        {t("settings.browse")}
                    </button>
                </div>
            </label>

            {binStatus && (
                <div className="bin-status">
                    <div className={binStatus.YtDlpFound ? "bin-ok" : "bin-missing"}>
                        yt-dlp: {binStatus.YtDlpFound ? binStatus.YtDlpPath : t("settings.notFound")}
                    </div>
                    <div className={binStatus.FFmpegFound ? "bin-ok" : "bin-missing"}>
                        ffmpeg: {binStatus.FFmpegFound ? binStatus.FFmpegPath : t("settings.notFound")}
                    </div>
                </div>
            )}

            <label className="settings-field">
                <span>{t("settings.cookieSource")}</span>
                <select
                    value={settings.Cookies.Kind}
                    onChange={(e) => set("Cookies", { ...settings.Cookies, Kind: e.target.value })}
                >
                    <option value="none">{t("settings.cookieNone")}</option>
                    <option value="browser">{t("settings.cookieBrowser")}</option>
                    <option value="file">{t("settings.cookieFile")}</option>
                </select>
            </label>

            {settings.Cookies.Kind === "browser" && (
                <label className="settings-field">
                    <span>{t("settings.browser")}</span>
                    <select
                        value={settings.Cookies.Browser}
                        onChange={(e) => set("Cookies", { ...settings.Cookies, Browser: e.target.value })}
                    >
                        <option value="chrome">Chrome</option>
                        <option value="edge">Edge</option>
                        <option value="firefox">Firefox</option>
                    </select>
                </label>
            )}

            {settings.Cookies.Kind === "file" && (
                <label className="settings-field">
                    <span>{t("settings.cookiesFilePath")}</span>
                    <div className="settings-path-row">
                        <input
                            value={settings.Cookies.FilePath}
                            onChange={(e) => set("Cookies", { ...settings.Cookies, FilePath: e.target.value })}
                        />
                        <button
                            onClick={async () => {
                                const path = await ChooseFile(t("settings.chooseCookies"));
                                if (path) set("Cookies", { ...settings.Cookies, FilePath: path });
                            }}
                        >
                            {t("settings.browse")}
                        </button>
                    </div>
                </label>
            )}

            <label className="settings-field">
                <span>{t("settings.theme")}</span>
                <select
                    value={settings.Theme}
                    onChange={(e) => {
                        set("Theme", e.target.value);
                        document.documentElement.dataset.theme = e.target.value;
                    }}
                >
                    <option value="dark">{t("settings.themeDark")}</option>
                    <option value="light">{t("settings.themeLight")}</option>
                </select>
            </label>

            <label className="settings-field">
                <span>{t("settings.language")}</span>
                <select
                    value={settings.Language}
                    onChange={(e) => {
                        set("Language", e.target.value);
                        if (isLanguage(e.target.value)) setLang(e.target.value);
                    }}
                >
                    {LANGUAGES.map((l) => (
                        <option key={l.code} value={l.code}>
                            {l.label}
                        </option>
                    ))}
                </select>
            </label>

            <button className="save-btn" onClick={handleSave} disabled={saving}>
                {saving ? t("settings.saving") : t("settings.save")}
            </button>
            {message && <div className="settings-message">{message}</div>}
        </div>
    );
}

export default SettingsPanel;
