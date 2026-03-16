import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { useTranslation } from "@/lib/I18nContext";
import api from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, Check, AlertCircle, Palette } from "lucide-react";

interface ThemeInfo {
  id: string;
  name: string;
  author: string;
  version: string;
  description: string;
  url: string;
}

export function ThemePage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { t } = useTranslation();

  const [themes, setThemes] = useState<ThemeInfo[]>([]);
  const [activeTheme, setActiveTheme] = useState<string>("default");
  const [loading, setLoading] = useState(true);
  const [applying, setApplying] = useState<string | null>(null);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    if (user && user.role !== "admin" && user.role !== "super_admin") {
      navigate("/");
      return;
    }

    loadData();
  }, [user, navigate]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [themesRes, configRes] = await Promise.all([
        api.get<{ themes: ThemeInfo[] }>("/themes"),
        api.get<{ activeTheme: string }>("/config/theme", {
          params: { _t: Date.now() }, // 加时间戳，破除缓存
        }),
      ]);
      setThemes(themesRes.data.themes || []);
      setActiveTheme(configRes.data.activeTheme || "default");
      console.log("activeTheme from API:", configRes.data.activeTheme);
    } catch (error) {
      console.error("加载主题数据失败:", error);
      setMessage({ type: "error", text: t('themePage.loadFailed') });
    } finally {
      setLoading(false);
    }
  };

  const handleApplyTheme = async (themeId: string) => {
    setApplying(themeId);
    setMessage(null);
    try {
      await api.put("/config/theme", { activeTheme: themeId });
      setActiveTheme(themeId);
      const themeName = themes.find((t) => t.id === themeId)?.name || themeId;
      setMessage({
        type: "success",
        text: t('themePage.applySuccess', { themeName }),
      });

      // Reload the page after a short delay so the new theme takes effect
      // setTimeout(() => {
      //   window.location.reload();
      // }, 1500);
    } catch (error) {
      console.error("应用主题失败:", error);
      setMessage({ type: "error", text: t('themePage.applyFailed') });
    } finally {
      setApplying(null);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Palette className="w-8 h-8" />
          {t('themePage.title')}
        </h1>
        <p className="text-gray-500 mt-2">
          {t('themePage.description')}
        </p>
      </div>

      {message && (
        <Alert
          className={
            message.type === "success"
              ? "border-green-200 bg-green-50"
              : "border-red-200 bg-red-50"
          }
        >
          {message.type === "success" ? (
            <Check className="w-4 h-4 text-green-600" />
          ) : (
            <AlertCircle className="w-4 h-4 text-red-600" />
          )}
          <AlertDescription
            className={
              message.type === "success" ? "text-green-800" : "text-red-800"
            }
          >
            {message.text}
          </AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4">
        {themes.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-center text-gray-500">{t('themePage.noThemesFound')}</p>
            </CardContent>
          </Card>
        ) : (
          themes.map((theme) => (
            <Card
              key={theme.id}
              className={`transition-all ${
                activeTheme === theme.id
                  ? "border-blue-500 border-2 bg-blue-50"
                  : "hover:border-gray-400"
              }`}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <CardTitle className="flex items-center gap-2 flex-wrap">
                      {theme.name}
                      {activeTheme === theme.id && (
                        <Badge className="bg-blue-500">{t('themePage.currentBadge')}</Badge>
                      )}
                    </CardTitle>
                    <CardDescription>
                      <div className="mt-2 space-y-1 text-sm">
                        <p>
                          <span className="font-semibold">{t('themePage.author')}:</span>{" "}
                          {theme.author}
                        </p>
                        <p>
                          <span className="font-semibold">{t('themePage.version')}:</span>{" "}
                          {theme.version}
                        </p>
                        {theme.description && (
                          <p>
                            <span className="font-semibold">{t('themePage.themeDescription')}:</span>{" "}
                            {theme.description}
                          </p>
                        )}
                        {theme.url && (
                          <p>
                            <span className="font-semibold">{t('themePage.link')}:</span>{" "}
                            <a
                              href={theme.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-blue-500 hover:underline break-all"
                            >
                              {theme.url}
                            </a>
                          </p>
                        )}
                      </div>
                    </CardDescription>
                  </div>
                  <Button
                    onClick={() => handleApplyTheme(theme.id)}
                    disabled={activeTheme === theme.id || applying !== null}
                    variant={activeTheme === theme.id ? "secondary" : "default"}
                    className="shrink-0"
                  >
                    {applying === theme.id && (
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    )}
                    {activeTheme === theme.id ? t('themePage.applied') : t('themePage.applyTheme')}
                  </Button>
                </div>
              </CardHeader>
            </Card>
          ))
        )}
      </div>

      <Card className="bg-blue-50 border-blue-200">
        <CardHeader>
          <CardTitle className="text-blue-900">{t('themePage.installationInstructions')}</CardTitle>
        </CardHeader>
        <CardContent className="text-blue-800 space-y-2">
          <p>
            1. {t('themePage.instruction1')}{" "}
            <code className="bg-white px-2 py-1 rounded">./data/theme/</code>{" "}
            {t('themePage.instruction2')}
          </p>
          <p>
            2. {t('themePage.instruction3')}{" "}
            <code className="bg-white px-2 py-1 rounded">vexgo-theme.json</code>{" "}
            {t('themePage.instruction4')}
          </p>
          <p>
            3. {t('themePage.instruction5')}{" "}
            <code className="bg-white px-2 py-1 rounded">dist/</code>{" "}
            {t('themePage.instruction6')}
          </p>
          <p>4. {t('themePage.instruction7')}</p>
        </CardContent>
      </Card>
    </div>
  );
}
