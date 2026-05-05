"use client";

import { useEffect, useState, useTransition } from "react";
import { CheckCircle2, Eye, EyeOff, Loader2, Plug, ShieldAlert, ShoppingBag, Truck, MessageCircle } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  fetchWAHASessions,
  testIntegration,
  updateFrenetIntegration,
  updateWAHAIntegration,
  updateWooCommerceIntegration,
} from "@/lib/api";
import type {
  FrenetIntegration,
  IntegrationsResponse,
  TestResult,
  WAHAIntegration,
  WAHASession,
  WooCommerceIntegration,
} from "@/lib/types";

export function IntegrationsBoard({ initial }: { initial: IntegrationsResponse }) {
  const [data, setData] = useState(initial);

  return (
    <div className="space-y-6">
      <WooCommerceCard
        value={data.woocommerce}
        onSaved={(next) => setData(next)}
      />
      <FrenetCard
        value={data.frenet}
        onSaved={(next) => setData(next)}
      />
      <WAHACard
        value={data.waha}
        onSaved={(next) => setData(next)}
      />
    </div>
  );
}

// =============================================================================
// WooCommerce
// =============================================================================
function WooCommerceCard({
  value,
  onSaved,
}: {
  value: WooCommerceIntegration;
  onSaved: (next: IntegrationsResponse) => void;
}) {
  const [form, setForm] = useState(value);
  const [pending, start] = useTransition();
  const [test, setTest] = useState<TestResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  function patch(p: Partial<WooCommerceIntegration>) {
    setForm({ ...form, ...p });
  }

  function save() {
    setError(null);
    start(async () => {
      try {
        const next = await updateWooCommerceIntegration(form);
        onSaved(next);
        setForm(next.woocommerce);
        setTest({ ok: true, message: "Salvo." });
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao salvar");
      }
    });
  }

  function runTest() {
    setError(null);
    start(async () => {
      try {
        const next = await updateWooCommerceIntegration(form);
        onSaved(next);
        setForm(next.woocommerce);
        const r = await testIntegration("woocommerce");
        setTest(r);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao testar");
      }
    });
  }

  return (
    <ProviderCard
      icon={<ShoppingBag className="size-4" strokeWidth={1.75} />}
      title="WooCommerce"
      subtitle="Loja de origem dos pedidos."
      configured={value.configured}
      enabled={form.enabled}
      onToggle={(v) => patch({ enabled: v })}
    >
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="URL da loja"
          placeholder="https://lizzon.com.br"
          value={form.url}
          onChange={(v) => patch({ url: v })}
        />
        <Field
          label="Consumer Key"
          placeholder="ck_..."
          value={form.consumer_key}
          mono
          onChange={(v) => patch({ consumer_key: v })}
        />
        <SecretField
          label="Consumer Secret"
          placeholder="cs_..."
          value={form.consumer_secret}
          onChange={(v) => patch({ consumer_secret: v })}
        />
        <SecretField
          label="Webhook Secret"
          value={form.webhook_secret}
          onChange={(v) => patch({ webhook_secret: v })}
          hint="Gere com `openssl rand -hex 32` e cole aqui + no WP."
        />
      </div>
      <Footer
        pending={pending}
        test={test}
        error={error}
        onTest={runTest}
        onSave={save}
      />
    </ProviderCard>
  );
}

// =============================================================================
// Frenet
// =============================================================================
function FrenetCard({
  value,
  onSaved,
}: {
  value: FrenetIntegration;
  onSaved: (next: IntegrationsResponse) => void;
}) {
  const [form, setForm] = useState(value);
  const [pending, start] = useTransition();
  const [test, setTest] = useState<TestResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  function patch(p: Partial<FrenetIntegration>) {
    setForm({ ...form, ...p });
  }

  function save() {
    setError(null);
    start(async () => {
      try {
        const next = await updateFrenetIntegration(form);
        onSaved(next);
        setForm(next.frenet);
        setTest({ ok: true, message: "Salvo." });
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao salvar");
      }
    });
  }

  function runTest() {
    setError(null);
    start(async () => {
      try {
        const next = await updateFrenetIntegration(form);
        onSaved(next);
        setForm(next.frenet);
        const r = await testIntegration("frenet");
        setTest(r);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao testar");
      }
    });
  }

  return (
    <ProviderCard
      icon={<Truck className="size-4" strokeWidth={1.75} />}
      title="Frenet"
      subtitle="Rastreamento e tabelas de prazo das transportadoras."
      configured={value.configured}
      enabled={form.enabled}
      onToggle={(v) => patch({ enabled: v })}
    >
      <div className="grid gap-4 md:grid-cols-2">
        <SecretField
          label="API Token"
          value={form.api_token}
          onChange={(v) => patch({ api_token: v })}
          hint="Token da API oficial — usado pra rastrear shipments."
        />
        <Field
          label="Email do painel"
          placeholder="contato@univerbeauty.com.br"
          value={form.panel_email}
          onChange={(v) => patch({ panel_email: v })}
          hint="Usado pelo auto-linker (scraper do painel Frenet)."
        />
        <SecretField
          label="Senha do painel"
          value={form.panel_password}
          onChange={(v) => patch({ panel_password: v })}
        />
      </div>
      <Footer
        pending={pending}
        test={test}
        error={error}
        onTest={runTest}
        onSave={save}
      />
    </ProviderCard>
  );
}

// =============================================================================
// WAHA
// =============================================================================
function WAHACard({
  value,
  onSaved,
}: {
  value: WAHAIntegration;
  onSaved: (next: IntegrationsResponse) => void;
}) {
  const [form, setForm] = useState(value);
  const [pending, start] = useTransition();
  const [test, setTest] = useState<TestResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sessions, setSessions] = useState<WAHASession[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(false);

  function patch(p: Partial<WAHAIntegration>) {
    setForm({ ...form, ...p });
  }

  async function loadSessions() {
    setSessionsLoading(true);
    try {
      const r = await fetchWAHASessions();
      setSessions(r.sessions ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Falha ao listar sessões");
    } finally {
      setSessionsLoading(false);
    }
  }

  // Lazy-load when the integration is configured so users don't pay for a
  // gateway round-trip if they're just opening the page to read the URL.
  useEffect(() => {
    if (value.configured) loadSessions();
  }, [value.configured]);

  function save() {
    setError(null);
    start(async () => {
      try {
        const next = await updateWAHAIntegration(form);
        onSaved(next);
        setForm(next.waha);
        setTest({ ok: true, message: "Salvo." });
        loadSessions();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao salvar");
      }
    });
  }

  function runTest() {
    setError(null);
    start(async () => {
      try {
        const next = await updateWAHAIntegration(form);
        onSaved(next);
        setForm(next.waha);
        const r = await testIntegration("waha");
        setTest(r);
        loadSessions();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao testar");
      }
    });
  }

  return (
    <ProviderCard
      icon={<MessageCircle className="size-4" strokeWidth={1.75} />}
      title="WhatsApp (WAHA)"
      subtitle="Notificações ao cliente em cada marco da entrega."
      configured={value.configured}
      enabled={form.enabled}
      onToggle={(v) => patch({ enabled: v })}
    >
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="URL do gateway"
          placeholder="https://whatsapp.univerzap.cloud"
          value={form.url}
          onChange={(v) => patch({ url: v })}
        />
        <SecretField
          label="API Key"
          value={form.api_key}
          onChange={(v) => patch({ api_key: v })}
          hint="Header X-Api-Key enviado em cada chamada."
        />
      </div>

      <div className="mt-4">
        <label className="block text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Sessão padrão
        </label>
        <div className="mt-1 flex items-center gap-2">
          <select
            value={form.default_session ?? ""}
            onChange={(e) => patch({ default_session: e.target.value })}
            disabled={!value.configured || sessionsLoading}
            className="w-full rounded-lg border border-input bg-background px-3 py-1.5 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
          >
            <option value="">— sem sessão padrão (escolher por envio) —</option>
            {sessions.map((s) => (
              <option key={s.name} value={s.name}>
                {s.name}
                {s.status ? ` · ${s.status.toLowerCase()}` : ""}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={loadSessions}
            disabled={!value.configured || sessionsLoading}
            className="rounded-md border border-border/60 px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted disabled:opacity-50"
          >
            {sessionsLoading ? "…" : "Atualizar"}
          </button>
        </div>
        <p className="mt-1 text-[11px] text-muted-foreground">
          Sessão usada por padrão pelos triggers automáticos e pelo botão
          “Notificar via WhatsApp”. Pode ser sobrescrita por envio.
        </p>
      </div>

      <Footer
        pending={pending}
        test={test}
        error={error}
        onTest={runTest}
        onSave={save}
      />
    </ProviderCard>
  );
}

// =============================================================================
// Building blocks
// =============================================================================
function ProviderCard({
  icon,
  title,
  subtitle,
  configured,
  enabled,
  onToggle,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  subtitle: string;
  configured: boolean;
  enabled: boolean;
  onToggle: (next: boolean) => void;
  children: React.ReactNode;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0 gap-3">
        <div className="flex items-start gap-3">
          <div className="mt-0.5 flex size-9 items-center justify-center rounded-lg bg-secondary text-foreground">
            {icon}
          </div>
          <div>
            <CardTitle className="flex items-center gap-2 text-base">
              {title}
              {configured ? (
                <span className="inline-flex items-center gap-1 rounded-full bg-success/10 px-2 py-0.5 text-[10px] font-medium text-success">
                  <CheckCircle2 className="size-3" />
                  Configurada
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 rounded-full bg-warning/10 px-2 py-0.5 text-[10px] font-medium text-warning">
                  <ShieldAlert className="size-3" />
                  Não configurada
                </span>
              )}
            </CardTitle>
            <CardDescription className="mt-1">{subtitle}</CardDescription>
          </div>
        </div>
        <label className="inline-flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(e) => onToggle(e.target.checked)}
            className="size-4 rounded border-border accent-foreground"
          />
          Habilitada
        </label>
      </CardHeader>
      <CardContent className="space-y-5">{children}</CardContent>
    </Card>
  );
}

function Field({
  label,
  value,
  onChange,
  placeholder,
  hint,
  mono,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  hint?: string;
  mono?: boolean;
}) {
  return (
    <label className="block">
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
      <input
        type="text"
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className={
          "mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring " +
          (mono ? "font-mono text-xs" : "")
        }
      />
      {hint ? <div className="mt-1 text-[11px] text-muted-foreground">{hint}</div> : null}
    </label>
  );
}

function SecretField({
  label,
  value,
  onChange,
  placeholder,
  hint,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  hint?: string;
}) {
  const [reveal, setReveal] = useState(false);
  return (
    <label className="block">
      <div className="text-xs font-medium text-muted-foreground">{label}</div>
      <div className="relative">
        <input
          type={reveal ? "text" : "password"}
          value={value}
          placeholder={placeholder}
          onChange={(e) => onChange(e.target.value)}
          className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 pr-10 font-mono text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
        <button
          type="button"
          onClick={() => setReveal((r) => !r)}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted"
          aria-label={reveal ? "Ocultar" : "Mostrar"}
        >
          {reveal ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
        </button>
      </div>
      {hint ? <div className="mt-1 text-[11px] text-muted-foreground">{hint}</div> : null}
    </label>
  );
}

function Footer({
  pending,
  test,
  error,
  onTest,
  onSave,
}: {
  pending: boolean;
  test: TestResult | null;
  error: string | null;
  onTest: () => void;
  onSave: () => void;
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
      <div className="min-h-[20px] text-xs">
        {error ? (
          <span className="text-destructive">{error}</span>
        ) : test ? (
          test.ok ? (
            <span className="inline-flex items-center gap-1 text-success">
              <CheckCircle2 className="size-3" /> {test.message ?? "OK"}
            </span>
          ) : (
            <span className="text-destructive">{test.error ?? "Falhou"}</span>
          )
        ) : null}
      </div>
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" onClick={onTest} disabled={pending}>
          {pending ? <Loader2 className="size-3.5 animate-spin" /> : <Plug className="size-3.5" />}
          Testar conexão
        </Button>
        <Button size="sm" onClick={onSave} disabled={pending}>
          {pending ? <Loader2 className="size-3.5 animate-spin" /> : null}
          Salvar
        </Button>
      </div>
    </div>
  );
}
