<script lang="ts">
  import { onMount } from 'svelte'
  import { debate } from './lib/debate.svelte'
  import Countdown from './lib/Countdown.svelte'
  import ChatMessage from './lib/ChatMessage.svelte'
  import TypingIndicator from './lib/TypingIndicator.svelte'

  let scroller = $state<HTMLElement>()

  onMount(() => debate.connect())

  // Split the roster into the two political sides so the UI can show clearly
  // who governs (Tidöavtalet) and who is in opposition.
  let tido = $derived(debate.bots.filter((b) => b.bloc === 'tido'))
  let opposition = $derived(debate.bots.filter((b) => b.bloc !== 'tido'))

  // Lightweight live metrics for the header stat cluster.
  let participantCount = $derived(debate.bots.length)
  let postCount = $derived(debate.messages.length)

  // Keep the chat pinned to the bottom as new messages / typing arrive.
  $effect(() => {
    // Touch reactive deps so this re-runs on change.
    void debate.messages.length
    void debate.typingBotId
    if (scroller) scroller.scrollTop = scroller.scrollHeight
  })
</script>

<main>
  <header>
    <div class="banner-wrap">
      <img
        class="banner"
        src="/banner.jpeg"
        alt="AI Valdebatten — Sveriges största digitala valdebatt"
      />
      <div class="banner-fade"></div>
    </div>
    <div class="statusbar">
      <div class="brand">
        <span class="brand-mark">AV</span>
        <span class="brand-name">AI&nbsp;Valdebatten</span>
      </div>
      <span class="status" class:on={debate.connected}>
        {debate.connected ? 'Live' : 'Offline'}
      </span>
    </div>
    <div class="head-text">
      <div class="eyebrow">Nuvarande ämne</div>
      <div class="topic">
        {debate.topic || 'Väntar på nästa debatt…'}
      </div>
      <div class="stats">
        <div class="stat">
          <span class="stat-value">{participantCount}</span>
          <span class="stat-label">Deltagare</span>
        </div>
        <div class="stat">
          <span class="stat-value">{postCount}</span>
          <span class="stat-label">Inlägg</span>
        </div>
        <div class="stat stat-timer">
          <span class="stat-value"><Countdown endsAt={debate.endsAt} /></span>
          <span class="stat-label">Nästa ämne</span>
        </div>
      </div>
    </div>
  </header>

  {#if debate.bots.length}
    <section class="roster">
      <div class="bloc bloc-tido">
        <h2 class="bloc-title">
          <span class="bloc-dot"></span>Tidöavtalet
          <span class="bloc-count">{tido.length}</span>
          <span class="bloc-sub">Regeringen</span>
        </h2>
        <div class="chips">
          {#each tido as b (b.id)}
            <span class="chip" style:--c={b.color} title={b.manifesto}>{b.name}</span>
          {/each}
        </div>
      </div>
      <div class="bloc bloc-opposition">
        <h2 class="bloc-title">
          <span class="bloc-dot"></span>Oppositionen
          <span class="bloc-count">{opposition.length}</span>
          <span class="bloc-sub">Utanför regeringen</span>
        </h2>
        <div class="chips">
          {#each opposition as b (b.id)}
            <span class="chip" style:--c={b.color} title={b.manifesto}>{b.name}</span>
          {/each}
        </div>
      </div>
    </section>
  {/if}

  <section class="chat" bind:this={scroller}>
    {#if debate.messages.length === 0 && !debate.typingBot}
      <p class="empty">Debatten startar strax…</p>
    {/if}
    {#each debate.messages as m (m.id)}
      <ChatMessage message={m} bot={debate.bot(m.botId)} />
    {/each}
    {#if debate.typingBot}
      <TypingIndicator bot={debate.typingBot} />
    {/if}
  </section>
</main>

<style>
  main {
    width: min(1040px, calc(100% - 2rem));
    height: min(100dvh - 2rem, 980px);
    margin: 1rem auto;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: color-mix(in srgb, var(--surface) 86%, transparent);
    border: 1px solid rgba(255, 255, 255, 0.78);
    border-radius: 28px;
    box-shadow: var(--shadow-soft);
    backdrop-filter: blur(18px);
  }

  header {
    position: relative;
    padding: 0 0 1rem;
    border-bottom: 1px solid rgba(0, 106, 167, 0.16);
    background:
      linear-gradient(145deg, rgba(255, 255, 255, 0.95), rgba(240, 247, 252, 0.9)),
      linear-gradient(90deg, var(--se-blue) 0 24%, var(--se-yellow) 24% 32%, var(--se-blue) 32%);
  }
  header::after {
    content: '';
    position: absolute;
    inset: auto 0 0;
    height: 4px;
    background: linear-gradient(90deg, var(--se-blue) 0 24%, var(--se-yellow) 24% 32%, var(--se-blue) 32%);
  }
  .banner-wrap {
    position: relative;
    overflow: hidden;
  }
  .banner {
    display: block;
    width: 100%;
    height: clamp(180px, 26vw, 260px);
    object-fit: cover;
    object-position: center 42%;
    transform: scale(1.01);
  }
  .banner-fade {
    position: absolute;
    inset: auto 0 0;
    height: 38%;
    pointer-events: none;
    background: linear-gradient(
      to bottom,
      rgba(255, 255, 255, 0) 0%,
      rgba(255, 255, 255, 0.18) 55%,
      rgba(255, 255, 255, 0.82) 100%
    );
  }
  .statusbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.9rem clamp(1rem, 2.5vw, 1.8rem) 0;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
  }
  .brand-mark {
    display: grid;
    place-items: center;
    width: 1.85rem;
    height: 1.85rem;
    border-radius: var(--radius-sm);
    font-family: var(--display-font);
    font-size: 0.78rem;
    font-weight: 800;
    letter-spacing: 0.02em;
    color: #fff;
    background: linear-gradient(150deg, var(--se-blue) 0%, var(--se-blue-dark) 100%);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.25), 0 6px 14px rgba(0, 63, 115, 0.22);
  }
  .brand-name {
    font-family: var(--display-font);
    font-size: 0.86rem;
    font-weight: 800;
    letter-spacing: -0.01em;
    color: var(--se-blue-ink);
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-family: var(--display-font);
    font-size: 0.66rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.13em;
    color: var(--text-muted);
    border: 1px solid rgba(99, 116, 135, 0.22);
    border-radius: 999px;
    padding: 0.26rem 0.66rem;
    background: rgba(255, 255, 255, 0.82);
    box-shadow: var(--shadow-xs);
  }
  .status::before {
    content: '';
    width: 0.45rem;
    height: 0.45rem;
    border-radius: 50%;
    background: #94a3b8;
  }
  .status.on {
    color: #15803d;
    border-color: rgba(34, 197, 94, 0.32);
    background: rgba(240, 253, 244, 0.92);
  }
  .status.on::before {
    background: #22c55e;
    box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.14);
    animation: pulse 2s ease-in-out infinite;
  }
  @keyframes pulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.32); }
    50% { box-shadow: 0 0 0 5px rgba(34, 197, 94, 0.06); }
  }
  .head-text {
    position: relative;
    padding: 0.75rem clamp(1rem, 2.5vw, 1.8rem) 0.45rem;
    margin: 0;
  }
  .head-text::before {
    content: '';
    position: absolute;
    left: clamp(1rem, 2.5vw, 1.8rem);
    top: 0.8rem;
    width: 3.7rem;
    height: 0.28rem;
    border-radius: 999px;
    background: var(--se-yellow);
  }
  .eyebrow {
    margin-top: 0.75rem;
    font-family: var(--display-font);
    font-size: 0.74rem;
    font-weight: 800;
    letter-spacing: 0.13em;
    text-transform: uppercase;
    color: var(--se-blue-dark);
  }
  .topic {
    max-width: 22ch;
    margin-top: 0.24rem;
    font-family: var(--display-font);
    font-size: clamp(1.35rem, 4vw, 2.45rem);
    font-weight: 800;
    letter-spacing: -0.045em;
    line-height: 0.96;
    color: var(--se-blue-ink);
  }
  .stats {
    display: flex;
    flex-wrap: wrap;
    gap: 0.55rem;
    margin-top: 0.85rem;
  }
  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.05rem;
    min-width: 4.5rem;
    padding: 0.45rem 0.85rem 0.5rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.78);
    box-shadow: var(--shadow-xs);
  }
  .stat-timer {
    border-color: rgba(0, 106, 167, 0.28);
    background: rgba(0, 106, 167, 0.06);
  }
  .stat-value {
    font-family: var(--display-font);
    font-size: 1.18rem;
    font-weight: 800;
    line-height: 1.05;
    letter-spacing: -0.02em;
    color: var(--se-blue-ink);
    font-variant-numeric: tabular-nums;
  }
  .stat-label {
    font-size: 0.62rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-muted);
  }

  @media (max-width: 640px) {
    .statusbar {
      padding: 0.65rem 0.9rem 0;
    }
    .head-text {
      padding: 0.65rem 0.9rem 0.35rem;
    }
    .head-text::before {
      left: 0.9rem;
    }
    .topic {
      max-width: 100%;
    }
  }

  @media (max-width: 560px) {
    .roster {
      grid-template-columns: 1fr;
    }
  }

  .roster {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
    padding: 0.85rem clamp(1rem, 2.5vw, 1.8rem);
    border-bottom: 1px solid rgba(0, 106, 167, 0.12);
    background: rgba(243, 247, 251, 0.76);
  }
  .bloc {
    border-radius: var(--radius-md);
    padding: 0.7rem 0.8rem 0.8rem;
    border: 1px solid transparent;
    box-shadow: var(--shadow-xs);
  }
  .bloc-tido {
    background: linear-gradient(180deg, rgba(0, 90, 168, 0.09), rgba(0, 90, 168, 0.04));
    border-color: rgba(0, 90, 168, 0.22);
  }
  .bloc-opposition {
    background: linear-gradient(180deg, rgba(237, 27, 52, 0.07), rgba(237, 27, 52, 0.03));
    border-color: rgba(237, 27, 52, 0.2);
  }
  .bloc-title {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    margin: 0 0 0.6rem;
    font-family: var(--display-font);
    font-size: 0.82rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--se-blue-ink);
  }
  .bloc-dot {
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    flex: 0 0 auto;
  }
  .bloc-tido .bloc-dot {
    background: #005ea8;
  }
  .bloc-opposition .bloc-dot {
    background: #ed1b34;
  }
  .bloc-count {
    display: inline-grid;
    place-items: center;
    min-width: 1.3rem;
    height: 1.3rem;
    padding: 0 0.35rem;
    border-radius: 999px;
    font-size: 0.68rem;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    background: rgba(255, 255, 255, 0.82);
    border: 1px solid rgba(8, 47, 73, 0.12);
    color: var(--se-blue-ink);
  }
  .bloc-sub {
    margin-left: auto;
    font-size: 0.62rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: var(--text-muted);
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .chip {
    font-size: 0.76rem;
    font-weight: 700;
    color: color-mix(in srgb, var(--c) 82%, #082f49);
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(255, 255, 255, 0.62)),
      color-mix(in srgb, var(--c) 12%, white);
    border: 1px solid color-mix(in srgb, var(--c) 42%, white);
    border-radius: 999px;
    padding: 0.26rem 0.72rem;
    box-shadow: 0 8px 20px rgba(0, 63, 115, 0.07);
    cursor: help;
  }

  .chat {
    flex: 1;
    overflow-y: auto;
    padding: clamp(1rem, 2.5vw, 1.65rem);
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    scroll-behavior: smooth;
    background:
      linear-gradient(rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.72)),
      repeating-linear-gradient(135deg, rgba(0, 106, 167, 0.055) 0 1px, transparent 1px 18px);
  }
  .empty {
    margin: auto;
    padding: 1rem 1.25rem;
    color: var(--text-muted);
    font-weight: 600;
    background: rgba(255, 255, 255, 0.74);
    border: 1px solid rgba(0, 106, 167, 0.12);
    border-radius: 18px;
  }

  @media (max-width: 720px) {
    main {
      width: 100%;
      height: 100dvh;
      margin: 0;
      border-width: 0;
      border-radius: 0;
    }
    .banner {
      height: clamp(140px, 38vw, 180px);
    }
  }
</style>
