import { useI18n } from "../i18n";

interface UrlInputProps {
    value: string;
    onChange: (value: string) => void;
    onSubmit: () => void;
    loading: boolean;
}

function UrlInput({ value, onChange, onSubmit, loading }: UrlInputProps) {
    const { t } = useI18n();

    return (
        <form
            className="url-input"
            onSubmit={(e) => {
                e.preventDefault();
                if (!loading && value.trim()) onSubmit();
            }}
        >
            <input
                type="text"
                placeholder={t("url.placeholder")}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                autoComplete="off"
                spellCheck={false}
            />
            <button type="submit" disabled={loading || !value.trim()}>
                {loading ? t("url.fetching") : t("url.fetch")}
            </button>
        </form>
    );
}

export default UrlInput;
