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
    <div class="brand">
      <span class="logo">⬢</span> aivaldebatten
      <span class="status" class:on={debate.connected}>
        {debate.connected ? 'live' : 'offline'}
      </span>
    </div>
    <div class="topic">
      {debate.topic || 'Waiting for the next debate…'}
    </div>
    <div class="meta">
      next topic in <Countdown endsAt={debate.endsAt} />
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
    width: min(720px, 100%);
    height: 100dvh;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    background: #0f172a;
  }

  header {
    padding: 0.9rem 1rem 0.8rem;
    border-bottom: 1px solid #1e293b;
    background: #0b1222;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    color: #e2e8f0;
  }
  .logo { color: #38bdf8; }
  .status {
    margin-left: auto;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: #64748b;
    border: 1px solid #334155;
    border-radius: 999px;
    padding: 0.05rem 0.5rem;
  }
  .status.on {
    color: #22c55e;
    border-color: #14532d;
  }
  .topic {
    margin-top: 0.55rem;
    font-size: 1.15rem;
    font-weight: 600;
    color: #f8fafc;
  }
  .meta {
    margin-top: 0.25rem;
    font-size: 0.8rem;
    color: #64748b;
  }

  .roster {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    padding: 0.6rem 1rem;
    border-bottom: 1px solid #1e293b;
  }
  .chip {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--c);
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
    color: #475569;
    font-style: italic;
  }
</style>
