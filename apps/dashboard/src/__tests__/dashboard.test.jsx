import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import Dashboard from '../dashboard.jsx';

// ── fetch mock helpers ────────────────────────────────────────────────────────

const makeStatusResponse = (overrides = {}) => ({
  mode: 'enterprise',
  version: 'v2.4.0',
  uptime: '3h 22m',
  queue_depth: 0,
  ...overrides,
});

const makeMetricsResponse = (overrides = {}) => ({
  requests_per_second: 10,
  pii_hits_per_type: { EMAIL: 5, PERSON: 3 },
  latency_per_tier: { regex: '8ms' },
  ...overrides,
});

// mockFetch installs a fetch spy that routes by URL substring.
function mockFetch(handlers = {}) {
  vi.spyOn(global, 'fetch').mockImplementation((url) => {
    if (String(url).includes('status') && handlers.status !== undefined) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(handlers.status) });
    }
    if (String(url).includes('metrics') && handlers.metrics !== undefined) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(handlers.metrics) });
    }
    if (String(url).includes('logs') && handlers.logs !== undefined) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(handlers.logs) });
    }
    if (String(url).includes('refine') && handlers.refine !== undefined) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(handlers.refine) });
    }
    if (String(url).includes('config') && handlers.config !== undefined) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(handlers.config) });
    }
    return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
  });
}

beforeEach(() => {
  vi.restoreAllMocks();
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ── MetricCard rendering ──────────────────────────────────────────────────────

describe('MetricCard (via Dashboard)', () => {
  it('renders all four metric card labels immediately', () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);

    expect(screen.getByText(/Requests \/ Hour/i)).toBeInTheDocument();
    expect(screen.getByText(/PII Detections/i)).toBeInTheDocument();
    expect(screen.getByText(/Refinery Latency/i)).toBeInTheDocument();
    expect(screen.getByText(/Worker Queue/i)).toBeInTheDocument();
  });

  it('displays fetched metric values after API response', async () => {
    mockFetch({
      status: makeStatusResponse(),
      metrics: makeMetricsResponse({ requests_per_second: 5, pii_hits_per_type: { EMAIL: 12 } }),
      logs: { logs: [] },
    });
    render(<Dashboard />);

    // 5 req/s × 3600 = 18,000
    await waitFor(() => expect(screen.getByText('18,000')).toBeInTheDocument(), { timeout: 3000 });
  });

  it('sums all pii_hits_per_type values for detection count', async () => {
    mockFetch({
      status: makeStatusResponse(),
      metrics: makeMetricsResponse({ pii_hits_per_type: { EMAIL: 7, SSN: 3 } }),
      logs: { logs: [] },
    });
    render(<Dashboard />);

    await waitFor(() => expect(screen.getByText('10')).toBeInTheDocument(), { timeout: 3000 });
  });
});

// ── System status ─────────────────────────────────────────────────────────────

describe('System status indicator', () => {
  it('shows "System Online" when status fetch succeeds', async () => {
    mockFetch({ status: makeStatusResponse(), logs: { logs: [] } });
    render(<Dashboard />);
    await waitFor(() => expect(screen.getByText(/System Online/i)).toBeInTheDocument(), { timeout: 3000 });
  });

  it('shows Enterprise edition label for enterprise mode', async () => {
    mockFetch({ status: makeStatusResponse({ mode: 'enterprise' }), logs: { logs: [] } });
    render(<Dashboard />);
    await waitFor(() => expect(screen.getByText(/Enterprise Edition/i)).toBeInTheDocument(), { timeout: 3000 });
  });

  it('shows Community edition label for non-enterprise mode', async () => {
    mockFetch({ status: makeStatusResponse({ mode: 'community' }), logs: { logs: [] } });
    render(<Dashboard />);
    await waitFor(() => expect(screen.getByText(/Community Edition/i)).toBeInTheDocument(), { timeout: 3000 });
  });

  it('shows "System Offline" when status fetch rejects', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue(new Error('network error'));
    render(<Dashboard />);
    await waitFor(() => expect(screen.getByText(/System Offline/i)).toBeInTheDocument(), { timeout: 3000 });
  });
});

// ── Tab navigation ─────────────────────────────────────────────────────────────

describe('Tab navigation', () => {
  it('renders overview tab content by default', () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);
    expect(screen.getByText(/Live Refinery Test/i)).toBeInTheDocument();
  });

  it('switches to config tab on click', async () => {
    mockFetch({ logs: { logs: [] }, config: { test: true } });
    render(<Dashboard />);

    fireEvent.click(screen.getByRole('button', { name: /config/i }));

    await waitFor(() => expect(screen.getByText(/Configuration Plane/i)).toBeInTheDocument(), { timeout: 3000 });
  });

  it('switches to logs tab on click', () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);

    fireEvent.click(screen.getByRole('button', { name: /logs/i }));
    expect(screen.getByText(/SYSTEM_JOURNAL/i)).toBeInTheDocument();
  });

  it('can navigate back to overview from another tab', () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);

    fireEvent.click(screen.getByRole('button', { name: /logs/i }));
    fireEvent.click(screen.getByRole('button', { name: /overview/i }));

    expect(screen.getByText(/Live Refinery Test/i)).toBeInTheDocument();
  });
});

// ── LogLine rendering ─────────────────────────────────────────────────────────

describe('LogLine (via audit trail)', () => {
  it('renders audit log entries in the overview trail', async () => {
    const logs = [
      { timestamp: new Date().toISOString(), action: 'REDACT', result: 'SUCCESS', notes: 'email masked' },
      { timestamp: new Date().toISOString(), action: 'TOKENIZE', result: 'SUCCESS', notes: 'ssn vaulted' },
    ];
    mockFetch({ status: makeStatusResponse(), logs: { logs }, metrics: makeMetricsResponse() });
    render(<Dashboard />);

    await waitFor(() => expect(screen.getByText(/email masked/i)).toBeInTheDocument(), { timeout: 3000 });
    expect(screen.getByText(/ssn vaulted/i)).toBeInTheDocument();
  });

  it('shows placeholder text when audit log is empty', async () => {
    mockFetch({ status: makeStatusResponse(), logs: { logs: [] }, metrics: makeMetricsResponse() });
    render(<Dashboard />);

    await waitFor(() => expect(screen.getByText(/No audit records/i)).toBeInTheDocument(), { timeout: 3000 });
  });
});

// ── Live Refinery Test panel ──────────────────────────────────────────────────

describe('Live Refinery Test panel', () => {
  it('execute button is disabled with empty input', () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);
    const btn = screen.getByRole('button', { name: /execute redaction/i });
    expect(btn).toBeDisabled();
  });

  it('execute button enables after user types input', async () => {
    mockFetch({ logs: { logs: [] } });
    render(<Dashboard />);

    const textarea = screen.getByPlaceholderText(/Paste logs/i);
    await userEvent.type(textarea, 'test text');

    expect(screen.getByRole('button', { name: /execute redaction/i })).not.toBeDisabled();
  });

  it('shows redacted output returned by the refinery', async () => {
    mockFetch({
      logs: { logs: [] },
      refine: { refined: 'Hello, my email is [EMAIL_a1b2c3d4]' },
    });
    render(<Dashboard />);

    const textarea = screen.getByPlaceholderText(/Paste logs/i);
    await userEvent.type(textarea, 'Hello, my email is test@example.com');

    fireEvent.click(screen.getByRole('button', { name: /execute redaction/i }));

    await waitFor(() => {
      expect(screen.getByText(/\[EMAIL_a1b2c3d4\]/)).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('shows error message when refinery is unreachable', async () => {
    vi.spyOn(global, 'fetch').mockImplementation((url) => {
      if (String(url).includes('refine')) return Promise.reject(new Error('network error'));
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ logs: [] }) });
    });
    render(<Dashboard />);

    const textarea = screen.getByPlaceholderText(/Paste logs/i);
    await userEvent.type(textarea, 'some input text');
    fireEvent.click(screen.getByRole('button', { name: /execute redaction/i }));

    await waitFor(() => {
      expect(screen.getByText(/ERROR.*Refinery unreachable/i)).toBeInTheDocument();
    }, { timeout: 3000 });
  });
});
