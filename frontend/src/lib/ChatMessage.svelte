<script lang="ts">
  import type { Bot, Message } from './types'

  let { message, bot }: { message: Message; bot: Bot | undefined } = $props()

  let color = $derived(bot?.color ?? '#888')
  let name = $derived(bot?.name ?? message.botId)
  let initial = $derived(name.slice(0, 1).toUpperCase())
  let isTido = $derived(bot?.bloc === 'tido')
  let blocLabel = $derived(isTido ? 'Tidöavtalet' : 'Opposition')
</script>

<div class="row" class:tido={isTido} class:opposition={!isTido}>
  <div class="avatar" style:--accent={color}>
    {#if bot?.avatarUrl}
      <img src={bot.avatarUrl} alt={name} />
    {:else}
      <span>{initial}</span>
    {/if}
  </div>
  <div class="bubble" style:--accent={color}>
    <div class="name" style:color={color}>
      {name}
      <span class="bloc-badge">{blocLabel}</span>
    </div>
    <div class="text">{message.text}</div>
  </div>
</div>

<style>
  .row {
    display: flex;
    gap: 0.72rem;
    align-items: flex-start;
    animation: pop 180ms ease-out;
  }
  .row.opposition {
    flex-direction: row-reverse;
  }
  .row.opposition .bubble {
    border-radius: 20px 6px 20px 20px;
  }
  .row.opposition .bubble::before {
    left: auto;
    right: 0;
  }
  .row.opposition .name {
    justify-content: flex-end;
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
    position: relative;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(248, 251, 253, 0.94)),
      color-mix(in srgb, var(--accent) 7%, white);
    border: 1px solid color-mix(in srgb, var(--accent) 22%, var(--border, #d3dde8));
    border-radius: 6px 20px 20px 20px;
    padding: 0.68rem 0.88rem 0.76rem;
    max-width: min(82%, 54ch);
    box-shadow: 0 12px 30px rgba(0, 63, 115, 0.08);
  }
  .bubble::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0.8rem;
    bottom: 0.8rem;
    width: 3px;
    border-radius: 999px;
    background: var(--accent);
  }
  .name {
    font-family: var(--display-font, system-ui);
    font-size: 0.78rem;
    font-weight: 800;
    letter-spacing: -0.01em;
    margin-bottom: 0.22rem;
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex-wrap: wrap;
  }
  .bloc-badge {
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 0.08rem 0.42rem;
    border-radius: 999px;
    border: 1px solid transparent;
  }
  .tido .bloc-badge {
    color: #004d8c;
    background: rgba(0, 90, 168, 0.12);
    border-color: rgba(0, 90, 168, 0.32);
  }
  .opposition .bloc-badge {
    color: #b0122a;
    background: rgba(237, 27, 52, 0.1);
    border-color: rgba(237, 27, 52, 0.3);
  }
  .text {
    color: var(--text, #1b2a38);
    line-height: 1.48;
    white-space: pre-wrap;
    word-wrap: break-word;
  }
  @media (max-width: 640px) {
    .bubble {
      max-width: calc(100% - 3.1rem);
    }
  }
  @keyframes pop {
    from { opacity: 0; transform: translateY(8px) scale(0.985); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
