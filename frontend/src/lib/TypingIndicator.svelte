<script lang="ts">
  import type { Bot } from './types'

  // Rendered while a bot is "typing" (between the bot_typing and message
  // events), giving the WhatsApp-style animated dots.
  let { bot }: { bot: Bot } = $props()

  let initial = $derived(bot.name.slice(0, 1).toUpperCase())
  let isTido = $derived(bot.bloc === 'tido')
</script>

<div class="row" class:opposition={!isTido}>
  <div class="avatar" style:--accent={bot.color}>
    {#if bot.avatarUrl}
      <img src={bot.avatarUrl} alt={bot.name} />
    {:else}
      <span>{initial}</span>
    {/if}
  </div>
  <div class="bubble" style:--accent={bot.color}>
    <span class="dot"></span>
    <span class="dot"></span>
    <span class="dot"></span>
  </div>
</div>

<style>
  .row {
    display: flex;
    gap: 0.72rem;
    align-items: flex-start;
  }
  .row.opposition {
    flex-direction: row-reverse;
  }
  .avatar {
    flex: 0 0 auto;
    width: 2.35rem;
    height: 2.35rem;
    border-radius: 50%;
    display: grid;
    place-items: center;
    overflow: hidden;
    background: var(--accent);
    color: #fff;
    font-family: var(--display-font, system-ui);
    font-weight: 800;
    font-size: 0.95rem;
    border: 2px solid rgba(255, 255, 255, 0.92);
    box-shadow: 0 10px 22px rgba(0, 63, 115, 0.14);
  }
  .avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .bubble {
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(248, 251, 253, 0.94)),
      color-mix(in srgb, var(--accent) 7%, white);
    border: 1px solid color-mix(in srgb, var(--accent) 22%, var(--border, #d3dde8));
    border-radius: 6px 20px 20px 20px;
    padding: 0.82rem 0.98rem;
    display: flex;
    gap: 0.3rem;
    align-items: center;
    box-shadow: 0 12px 30px rgba(0, 63, 115, 0.08);
  }
  .dot {
    width: 0.45rem;
    height: 0.45rem;
    border-radius: 50%;
    background: color-mix(in srgb, var(--accent) 45%, #94a3b8);
    animation: blink 1.2s infinite ease-in-out;
  }
  .dot:nth-child(2) { animation-delay: 0.2s; }
  .dot:nth-child(3) { animation-delay: 0.4s; }
  @keyframes blink {
    0%, 80%, 100% { opacity: 0.3; transform: translateY(0); }
    40% { opacity: 1; transform: translateY(-2px); }
  }
</style>
