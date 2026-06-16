<script lang="ts">
  import type { Bot, Message } from './types'

  let { message, bot }: { message: Message; bot: Bot | undefined } = $props()

  let color = $derived(bot?.color ?? '#888')
  let name = $derived(bot?.name ?? message.botId)
  let initial = $derived(name.slice(0, 1).toUpperCase())
</script>

<div class="row">
  <div class="avatar" style:background={color}>{initial}</div>
  <div class="bubble" style:--accent={color}>
    <div class="name" style:color={color}>{name}</div>
    <div class="text">{message.text}</div>
  </div>
</div>

<style>
  .row {
    display: flex;
    gap: 0.6rem;
    align-items: flex-end;
    animation: pop 160ms ease-out;
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
    background: #1e293b;
    border-left: 3px solid var(--accent);
    border-radius: 0 12px 12px 12px;
    padding: 0.5rem 0.75rem;
    max-width: min(80%, 46ch);
  }
  .name {
    font-size: 0.75rem;
    font-weight: 700;
    margin-bottom: 0.15rem;
  }
  .text {
    color: #e2e8f0;
    line-height: 1.4;
    white-space: pre-wrap;
    word-wrap: break-word;
  }
  @keyframes pop {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
