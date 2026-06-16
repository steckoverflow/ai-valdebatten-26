<script lang="ts">
  // Shows the time remaining in the current 15-minute cycle, ticking once a
  // second. Derived purely from endsAt so it stays correct across reconnects.
  let { endsAt }: { endsAt: Date | null } = $props()

  let now = $state(Date.now())

  $effect(() => {
    const id = setInterval(() => (now = Date.now()), 1000)
    return () => clearInterval(id)
  })

  let remainingMs = $derived(endsAt ? Math.max(0, endsAt.getTime() - now) : 0)

  function fmt(ms: number): string {
    const totalSec = Math.floor(ms / 1000)
    const m = Math.floor(totalSec / 60)
    const s = totalSec % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }
</script>

<span class="countdown" title="Time until next topic">{fmt(remainingMs)}</span>

<style>
  .countdown {
    font-variant-numeric: tabular-nums;
    font-weight: 700;
    color: var(--se-blue, #006aa7);
  }
</style>
