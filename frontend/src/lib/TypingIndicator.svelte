<script lang="ts">
  import type { Bot } from './types'

  // Rendered while a bot is "typing" (between the bot_typing and message
  // events), giving the WhatsApp-style animated dots.
  let { bot }: { bot: Bot } = $props()

  let initial = $derived(bot.name.slice(0, 1).toUpperCase())
</script>

<div class="row">
  <div class="avatar" style:background={bot.color}>{initial}</div>
  <div class="bubble" style:--accent={bot.color}>
    <span class="dot"></span>
    <span class="dot"></span>
    <span class="dot"></span>
  </div>
</div>

<style>
  .row {
    display: flex;
    gap: 0.6rem;
    align-items: flex-end;
  }
  .avatar {
    flex: 0 0 auto;
    width: 2.1rem;
    height: 2.1rem;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    font-weight: 700;
    font-size: 0.9rem;
  }
  .bubble {
    background: var(--surface-alt, #eef3f8);
    border: 1px solid var(--border, #d3dde8);
    border-left: 3px solid var(--accent);
    border-radius: 0 12px 12px 12px;
    padding: 0.7rem 0.85rem;
    display: flex;
    gap: 0.3rem;
    align-items: center;
  }
  .dot {
    width: 0.45rem;
    height: 0.45rem;
    border-radius: 50%;
    background: #94a3b8;
    animation: blink 1.2s infinite ease-in-out;
  }
  .dot:nth-child(2) { animation-delay: 0.2s; }
  .dot:nth-child(3) { animation-delay: 0.4s; }
  @keyframes blink {
    0%, 80%, 100% { opacity: 0.3; transform: translateY(0); }
    40% { opacity: 1; transform: translateY(-2px); }
  }
</style>
