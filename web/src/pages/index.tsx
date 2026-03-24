import {type ReactNode, useState, useEffect, useRef} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

/* ── Copy-to-clipboard install box ── */

function CopyBox({command, className}: {command: string; className?: string}) {
  const [copied, setCopied] = useState(false);

  function copy() {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <div className={clsx(styles.installBox, className)} onClick={copy} role="button" tabIndex={0}>
      <span className={styles.installPrompt}>$</span>
      <code className={styles.installCode}>{command}</code>
      <button className={styles.copyBtn} onClick={copy} aria-label="Copy to clipboard">
        {copied ? (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        ) : (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
        )}
      </button>
    </div>
  );
}

/* ── Terminal tab content panels ── */

function PanelPush() {
  return <>
    <div className={styles.termLine}>
      <span className={styles.termPrompt}>$</span>{' '}
      <span className={styles.termCmd}>kit push</span>{' '}
      <span className={styles.termArg}>./deploy-k8s</span>{' '}
      <span className={styles.termFlag}>--team backend</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>Resolving skill directory...</div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>Pushed</span> deploy-k8s to <span className={styles.termCyan}>backend</span> (v1)
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termDim}>sha256:a3f8c1...</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>Done.</span> 1 skill pushed.
    </div>
  </>;
}

function PanelInstall() {
  return <>
    <div className={styles.termLine}>
      <span className={styles.termPrompt}>$</span>{' '}
      <span className={styles.termCmd}>kit install</span>{' '}
      <span className={styles.termArg}>backend</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>Fetching namespace <span className={styles.termCyan}>backend</span>...</div>
    <div className={styles.termOutput}>Found 3 items (2 skills, 1 hook)</div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> deploy-k8s{' '}
      <span className={styles.termDim}>{'->'}</span> ~/.claude/commands/
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> deploy-k8s{' '}
      <span className={styles.termDim}>{'->'}</span> ~/.codex/
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> deploy-k8s{' '}
      <span className={styles.termDim}>{'->'}</span> ~/.cursor/rules/
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> lint-on-save{' '}
      <span className={styles.termDim}>{'->'}</span> ~/.claude/hooks/
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> review-pr{' '}
      <span className={styles.termDim}>{'->'}</span> ~/.claude/commands/
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>Done.</span> 3 items installed to 3 tools.
    </div>
  </>;
}

function PanelStatus() {
  return <>
    <div className={styles.termLine}>
      <span className={styles.termPrompt}>$</span>{' '}
      <span className={styles.termCmd}>kit status</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termDim}>ITEM{'\u00A0'.repeat(14)}VERSION{'\u00A0'.repeat(4)}STATUS</span>
    </div>
    <div className={styles.termOutput}>
      deploy-k8s{'\u00A0'.repeat(8)}v3{'\u00A0'.repeat(9)}
      <span className={styles.termGreen}>up to date</span>
    </div>
    <div className={styles.termOutput}>
      review-pr{'\u00A0'.repeat(9)}v1{'\u00A0'.repeat(9)}
      <span className={styles.termYellow}>v2 available</span>
    </div>
    <div className={styles.termOutput}>
      lint-on-save{'\u00A0'.repeat(6)}v2{'\u00A0'.repeat(9)}
      <span className={styles.termGreen}>up to date</span>
    </div>
    <div className={styles.termOutput}>
      db-migrate{'\u00A0'.repeat(8)}v1{'\u00A0'.repeat(9)}
      <span className={styles.termYellow}>v3 available</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termDim}>4 items installed, 2 updates available</span>
    </div>
  </>;
}

function PanelProfile() {
  return <>
    <div className={styles.termLine}>
      <span className={styles.termPrompt}>$</span>{' '}
      <span className={styles.termCmd}>kit install</span>{' '}
      <span className={styles.termFlag}>-p</span>{' '}
      <span className={styles.termArg}>backend/default</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>Loading profile <span className={styles.termCyan}>backend/default</span>...</div>
    <div className={styles.termOutput}>Profile contains 4 items:</div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> deploy-k8s{'\u00A0'.repeat(6)}
      <span className={styles.termDim}>skill</span>
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> review-pr{'\u00A0'.repeat(7)}
      <span className={styles.termDim}>skill</span>
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> lint-on-save{'\u00A0'.repeat(4)}
      <span className={styles.termDim}>hook</span>
    </div>
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>+</span> .claude-config{'\u00A0'.repeat(2)}
      <span className={styles.termDim}>config</span>
    </div>
    <div className={styles.termBlank} />
    <div className={styles.termOutput}>
      <span className={styles.termGreen}>Done.</span> Profile installed to Claude Code, Codex CLI, Cursor.
    </div>
  </>;
}

/* ── Terminal mockup ── */

const DEMO_TABS = [
  {label: 'push',    Panel: PanelPush},
  {label: 'install', Panel: PanelInstall},
  {label: 'status',  Panel: PanelStatus},
  {label: 'profile', Panel: PanelProfile},
];

function TerminalDemo({tabs, className}: {tabs: typeof DEMO_TABS; className?: string}) {
  const [active, setActive] = useState(0);
  const pauseRef = useRef(false);

  useEffect(() => {
    const id = setInterval(() => {
      if (!pauseRef.current) {
        setActive(i => (i + 1) % tabs.length);
      }
    }, 3200);
    return () => clearInterval(id);
  }, [tabs.length]);

  function pick(i: number) {
    setActive(i);
    pauseRef.current = true;
  }

  const {Panel} = tabs[active];

  return (
    <div className={clsx(styles.terminal, className)}>
      <div className={styles.terminalBar}>
        <span className={styles.terminalDot} style={{background: '#e06c75'}} />
        <span className={styles.terminalDot} style={{background: '#e5c07b'}} />
        <span className={styles.terminalDot} style={{background: '#98c379'}} />
        <div className={styles.demoTabs}>
          {tabs.map((t, i) => (
            <button key={t.label}
              className={clsx(styles.demoTab, i === active && styles.demoTabActive)}
              onClick={() => pick(i)}>
              {t.label}
            </button>
          ))}
        </div>
      </div>
      <div key={active} className={styles.terminalBody}>
        <Panel />
      </div>
    </div>
  );
}

/* ── Static terminal (no tabs) ── */

function TerminalStatic({children, title, className}: {children: ReactNode; title?: string; className?: string}) {
  return (
    <div className={clsx(styles.terminal, className)}>
      <div className={styles.terminalBar}>
        <span className={styles.terminalDot} style={{background: '#e06c75'}} />
        <span className={styles.terminalDot} style={{background: '#e5c07b'}} />
        <span className={styles.terminalDot} style={{background: '#98c379'}} />
        {title && <span className={styles.terminalTitle}>{title}</span>}
      </div>
      <div className={styles.terminalBody}>
        {children}
      </div>
    </div>
  );
}

/* ── Features ── */

const FEATURES = [
  {
    title: 'Multi-tool targeting',
    desc: 'Install to Claude Code, Codex CLI, and Cursor simultaneously. Auto-detects what\'s installed.',
    code: 'kit install backend --target claude\nkit install backend --target codex\nkit install backend  # all tools',
  },
  {
    title: 'Personal encryption',
    desc: 'Your personal namespace is AES-256-GCM encrypted at rest. Key derived per-user, never stored.',
    code: 'kit push ./secret-sauce\n# encrypted with HMAC(secret, email)\n# only you can decrypt',
  },
  {
    title: 'Team namespaces',
    desc: 'Push to shared team spaces. No setup — teams auto-create on first push.',
    code: 'kit push ./rules --team backend\nkit push ./hooks --team frontend\nkit list backend',
  },
  {
    title: 'Profiles',
    desc: 'Bundle skills, hooks, and configs into profiles. The "new laptop" command.',
    code: 'kit profile create my-setup\nkit profile add my-setup deploy-k8s\nkit install -p my-setup  # done',
  },
  {
    title: 'Self-hosted',
    desc: 'Docker container + Postgres. Your data, your server. No SaaS dependency.',
    code: 'docker compose up -d\n# That\'s it. Server on :8430',
  },
  {
    title: 'Stateless auth',
    desc: 'Email + JWT. No passwords, no OAuth, no user table. Token lives locally like SSH keys.',
    code: 'kit login --server https://kit.co\n> Email: you@company.com\n> Logged in.',
  },
];

/* ── CLI commands ── */

const COMMANDS = [
  {cmd: 'kit login',     desc: 'Connect to a kit server'},
  {cmd: 'kit push',      desc: 'Push skills, hooks, or configs'},
  {cmd: 'kit install',   desc: 'Install from server to local tools'},
  {cmd: 'kit uninstall', desc: 'Remove installed items'},
  {cmd: 'kit list',      desc: 'Browse items on the server'},
  {cmd: 'kit search',    desc: 'Search across team namespaces'},
  {cmd: 'kit status',    desc: 'Check what\'s installed and outdated'},
  {cmd: 'kit update',    desc: 'Pull latest versions'},
  {cmd: 'kit sync',      desc: 'Update everything installed'},
  {cmd: 'kit doctor',    desc: 'Health check'},
  {cmd: 'kit profile',   desc: 'Create and manage profiles'},
  {cmd: 'kit whoami',    desc: 'Show current identity'},
];

/* ── Page ── */

export default function Home(): ReactNode {
  return (
    <Layout title="kit — share AI coding skills, hooks, and config" description="Share AI coding skills, hooks, and config across tools and teams. Push once, install everywhere. Supports Claude Code, Codex CLI, and Cursor.">
    <div className={styles.pageWrap}>

      {/* ── Hero ── */}
      <section className={styles.hero}>
        <div className={styles.heroGlow} />
        <div className={styles.heroInner}>
          <div className={styles.heroLeft}>
            <div className={styles.heroBadge}>
              <span className={styles.heroBadgeText}>cli &middot; server &middot; open source</span>
            </div>
            <Heading as="h1" className={styles.heroTitle}>
              <span className={styles.heroGradient}>kit</span>
            </Heading>
            <p className={styles.heroTagline}>
              Share AI coding skills, hooks, and config across tools and teams.<br />
              Push once, install everywhere.
            </p>

            <div className={styles.heroInstall}>
              <CopyBox command="brew install dunkinfrunkin/tap/kit" />
              <span className={styles.installOr}>or</span>
              <CopyBox command="go install github.com/dunkinfrunkin/kit/cmd/kit@latest" />
            </div>

            <div className={styles.heroButtons}>
              <Link className={clsx('button button--primary button--md', styles.heroPrimary)}
                to="https://github.com/dunkinfrunkin/kit">
                GitHub
              </Link>
              <Link className={clsx('button button--secondary button--md', styles.heroSecondary)}
                to="#features">
                Docs
              </Link>
            </div>
          </div>

          <div className={styles.heroRight}>
            <TerminalDemo tabs={DEMO_TABS} />
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── Features ── */}
      <section id="features" className={styles.features}>
        <div className="container">
          <div className={styles.sectionHeader}>
            <span className={styles.sectionCaret}>$</span>
            <Heading as="h2" className={styles.sectionTitle}>Everything you need</Heading>
            <p className={styles.sectionSubtitle}>
              Push skills, hooks, and config to a server. Install them across Claude Code, Codex CLI, and Cursor in one command.
            </p>
          </div>
          <div className={styles.featuresGrid}>
            {FEATURES.map(f => (
              <div key={f.title} className={styles.featureCard}>
                <div className={styles.featureTitle}>{f.title}</div>
                <p className={styles.featureDesc}>{f.desc}</p>
                <pre className={styles.featureCode}>{f.code}</pre>
              </div>
            ))}
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── Architecture ── */}
      <section className={styles.architectureSection}>
        <div className="container">
          <div className={styles.sectionHeader}>
            <span className={styles.sectionCaret}>$</span>
            <Heading as="h2" className={styles.sectionTitle}>How it works</Heading>
            <p className={styles.sectionSubtitle}>
              A server holds namespaces. Developers push and pull. Tools stay in sync.
            </p>
          </div>
          <div className={styles.architectureWrap}>
            <TerminalStatic title="kit architecture">
              <div className={styles.archDiagram}>
{`┌─────────────────────────────────┐
│          Kit Server              │
│      (Docker + Postgres)         │
│                                  │
│  @frank/    personal (encrypted) │
│  @sarah/    personal (encrypted) │
│  backend/   team (plaintext)     │
│  frontend/  team (plaintext)     │
└─────────────────────────────────┘
        ▲              ▲
        │ push         │ pull
   ┌────┴────┐    ┌────┴────┐
   │  Dev A  │    │  Dev B  │
   └────┬────┘    └────┬────┘
        │              │
        ▼              ▼
   ~/.claude/      ~/.claude/
   ~/.codex/       ~/.codex/
   ~/.cursor/      ~/.cursor/`}
              </div>
            </TerminalStatic>
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── CLI commands ── */}
      <section className={styles.commandsSection}>
        <div className="container">
          <div className={styles.sectionHeader}>
            <span className={styles.sectionCaret}>$</span>
            <Heading as="h2" className={styles.sectionTitle}>CLI commands</Heading>
            <p className={styles.sectionSubtitle}>
              Everything is one command away.
            </p>
          </div>
          <div className={styles.keyGrid}>
            {COMMANDS.map(c => (
              <div key={c.cmd} className={styles.keyRow}>
                <div className={styles.keyBadge}>
                  <kbd className={styles.kbdKey}>{c.cmd}</kbd>
                </div>
                <div className={styles.keyAction}>{c.desc}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── New laptop ── */}
      <section className={styles.newLaptopSection}>
        <div className="container">
          <div className={styles.sectionHeader}>
            <span className={styles.sectionCaret}>$</span>
            <Heading as="h2" className={styles.sectionTitle}>New laptop? Three commands.</Heading>
            <p className={styles.sectionSubtitle}>
              Zero to fully equipped.
            </p>
          </div>
          <div className={styles.newLaptopWrap}>
            <TerminalStatic title="setup">
              <div className={styles.termLine}>
                <span className={styles.termPrompt}>$</span>{' '}
                <span className={styles.termCmd}>brew install</span>{' '}
                <span className={styles.termArg}>dunkinfrunkin/tap/kit</span>
              </div>
              <div className={styles.termBlank} />
              <div className={styles.termLine}>
                <span className={styles.termPrompt}>$</span>{' '}
                <span className={styles.termCmd}>kit login</span>{' '}
                <span className={styles.termFlag}>--server</span>{' '}
                <span className={styles.termArg}>https://kit.mycompany.com</span>
              </div>
              <div className={styles.termBlank} />
              <div className={styles.termLine}>
                <span className={styles.termPrompt}>$</span>{' '}
                <span className={styles.termCmd}>kit install</span>{' '}
                <span className={styles.termFlag}>-p</span>{' '}
                <span className={styles.termArg}>my-setup</span>
              </div>
              <div className={styles.termBlank} />
              <div className={styles.termOutput}>
                <span className={styles.termDim}># Done. Skills, hooks, configs installed across Claude, Codex, and Cursor.</span>
              </div>
            </TerminalStatic>
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── CTA ── */}
      <section className={styles.cta}>
        <div className={styles.ctaInner}>
          <Heading as="h2" className={styles.ctaTitle}>
            Try it <span className={styles.heroGradient}>right now</span>
          </Heading>
          <p className={styles.ctaSubtitle}>
            Install the CLI and start pushing skills in under a minute.
          </p>
          <CopyBox command="brew install dunkinfrunkin/tap/kit" className={styles.ctaBox} />
          <span className={styles.installOr}>or</span>
          <CopyBox command="go install github.com/dunkinfrunkin/kit/cmd/kit@latest" className={styles.ctaBox} />
          <div className={styles.ctaButtons}>
            <Link className={clsx('button button--primary button--lg', styles.heroPrimary)}
              to="https://github.com/dunkinfrunkin/kit">
              View on GitHub
            </Link>
            <Link className={clsx('button button--secondary button--lg', styles.heroSecondary)}
              to="#features">
              Read the docs
            </Link>
          </div>
        </div>
      </section>

      <div className={styles.divider} />

      {/* ── Follow ── */}
      <section className={styles.followSection}>
        <div className={styles.followInner}>
          <Heading as="h2" className={styles.followTitle}>Follow me</Heading>
          <div className={styles.followLinks}>
            <a href="https://github.com/dunkinfrunkin" target="_blank" rel="noopener noreferrer" className={styles.followLink}>
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
              <span>GitHub</span>
            </a>
            <a href="https://x.com/dunkinfrunkin" target="_blank" rel="noopener noreferrer" className={styles.followLink}>
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>
              <span>X</span>
            </a>
            <a href="https://www.linkedin.com/in/dunkinfrunkin/" target="_blank" rel="noopener noreferrer" className={styles.followLink}>
              <svg viewBox="0 0 24 24" width="23" height="23" fill="currentColor"><path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>
              <span>LinkedIn</span>
            </a>
          </div>
        </div>
      </section>

    </div>
    </Layout>
  );
}
