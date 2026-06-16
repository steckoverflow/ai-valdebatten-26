<script lang="ts">
  import { onMount } from 'svelte'
  import { debate } from './lib/debate.svelte'
  import Countdown from './lib/Countdown.svelte'
  import ChatMessage from './lib/ChatMessage.svelte'
  import TypingIndicator from './lib/TypingIndicator.svelte'

  let scroller = $state<HTMLElement>()

  onMount(() => debate.connect())

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
      <div class="topic">
        {debate.topic || 'Waiting for the next debate…'}
      </div>
      <div class="meta">
        next topic in <Countdown endsAt={debate.endsAt} />
      </div>
    </div>
  </header>

  {#if debate.bots.length}
    <section class="roster">
      {#each debate.bots as b (b.id)}
        <span class="chip" style:--c={b.color} title={b.manifesto}>{b.name}</span>
      {/each}
    </section>
  {/if}

  <section class="chat" bind:this={scroller}>
    {#if debate.messages.length === 0 && !debate.typingBot}
      <p class="empty">The debate is about to begin…</p>
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
    width: min(960px, 100%);
    height: 100dvh;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    background: var(--surface);
    box-shadow: 0 0 40px rgba(0, 69, 122, 0.08);
  }

  header {
    padding: 0;
    border-bottom: 3px solid var(--se-yellow);
    background: var(--surface);
  }
  .banner {
    display: block;
    width: 100%;
    height: auto;
  }
  .statusbar {
    display: flex;
    justify-content: flex-end;
    padding: 0.6rem 1rem 0;
  }
  .status {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 0.05rem 0.5rem;
  }
  .status.on {
    color: #15803d;
    border-color: #86efac;
    background: #f0fdf4;
  }
  .head-text {
    padding: 0.4rem 1rem 0.8rem;
    border-left: 4px solid var(--se-blue);
    margin: 0 1rem;
  }
  .topic {
    font-size: 1.15rem;
    font-weight: 700;
    color: var(--se-blue);
  }
  .meta {
    margin-top: 0.25rem;
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  @media (max-width: 640px) {
    .statusbar {
      padding: 0.45rem 0.8rem 0;
    }
    .head-text {
      padding: 0.35rem 0.8rem 0.6rem;
      margin: 0 0.8rem;
    }
    .topic {
      font-size: 1rem;
    }
  }

  .roster {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    padding: 0.6rem 1rem;
    border-bottom: 1px solid var(--border);
    background: var(--surface-alt);
  }
  .chip {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--c);
    background: var(--surface);
    border: 1px solid var(--c);
    border-radius: 999px;
    padding: 0.1rem 0.55rem;
    cursor: help;
  }

  .chat {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
    scroll-behavior: smooth;
  }
  .empty {
    margin: auto;
    color: var(--text-muted);
    font-style: italic;
  }
</style>
