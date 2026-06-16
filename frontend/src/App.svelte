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
    <img
      class="banner"
      src="/banner.png"
      alt="AI Valdebatten — Sveriges största digitala valdebatt"
    />
    <div class="statusbar">
      <span class="status" class:on={debate.connected}>
        {debate.connected ? 'live' : 'offline'}
      </span>
    </div>
    <div class="head-text">
      <div class="eyebrow">Sveriges digitala valdebatt</div>
      <div class="topic">
        {debate.topic || 'Väntar på nästa debatt…'}
      </div>
      <div class="meta">
        Nästa ämne om <Countdown endsAt={debate.endsAt} />
      </div>
    </div>
  </header>

  {#if debate.bots.length}
    <section class="roster">
      <div class="bloc bloc-tido">
        <h2 class="bloc-title">
          <span class="bloc-dot"></span>Tidöavtalet
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
  .banner {
    display: block;
    width: 100%;
    height: auto;
    max-height: 210px;
    object-fit: cover;
  }
  .statusbar {
    display: flex;
    justify-content: flex-end;
    padding: 0.9rem clamp(1rem, 2.5vw, 1.8rem) 0;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-family: var(--display-font);
    font-size: 0.68rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--text-muted);
    border: 1px solid rgba(99, 116, 135, 0.22);
    border-radius: 999px;
    padding: 0.24rem 0.64rem;
    background: rgba(255, 255, 255, 0.72);
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
    border-color: rgba(34, 197, 94, 0.28);
    background: rgba(240, 253, 244, 0.9);
  }
  .status.on::before {
    background: #22c55e;
    box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.14);
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
  .meta {
    margin-top: 0.65rem;
    font-size: 0.88rem;
    font-weight: 500;
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
    border-radius: 16px;
    padding: 0.7rem 0.8rem 0.8rem;
    border: 1px solid transparent;
  }
  .bloc-tido {
    background: rgba(0, 90, 168, 0.07);
    border-color: rgba(0, 90, 168, 0.22);
  }
  .bloc-opposition {
    background: rgba(237, 27, 52, 0.06);
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
      max-height: 150px;
    }
  }
</style>
