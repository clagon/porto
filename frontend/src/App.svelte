<script>
  import { onMount } from 'svelte';
  import { api } from './lib/api';
  import { busy, settings, status } from './lib/stores';
  import { validateSettings } from './lib/validate';

  let error = '';
  let form = {
    listen_addr: '127.0.0.1:8080',
    auto_discover: true,
  };

  function getPortCount() {
    return Array.isArray($status?.ports) ? $status.ports.length : 0;
  }

  async function refresh() {
    error = '';
    busy.set(true);
    try {
      const [nextStatus, nextSettings] = await Promise.all([api.status(), api.getSettings()]);
      status.set(nextStatus);
      settings.set(nextSettings);
      form = nextSettings;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      busy.set(false);
    }
  }

  async function runAction(action) {
    error = '';
    busy.set(true);
    try {
      await action();
      await refresh();
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      busy.set(false);
    }
  }

  async function discover() {
    await runAction(() => api.discover());
  }

  async function openPort() {
    await runAction(() => api.openPort());
  }

  async function closePort() {
    await runAction(() => api.closePort());
  }

  async function save() {
    const errors = validateSettings(form);
    if (errors.length > 0) {
      error = errors.join(', ');
      return;
    }
    await runAction(() => api.saveSettings(form));
  }

  onMount(refresh);
</script>

<main class="shell">
  <section class="panel hero">
    <div class="hero-copy">
      <p class="eyebrow">Local-only UPnP helper</p>
      <h1>port-mapper</h1>
      <p class="lede">Open and close router port mappings from a localhost-only UI. Configuration is saved next to the binary.</p>

      {#if error}
        <div class="alert" role="alert">
          {error}
        </div>
      {/if}

      <div class="actions">
        <button class="secondary" on:click={refresh} disabled={$busy}>Reload status</button>
        <button class="secondary" on:click={discover} disabled={$busy}>Re-scan router</button>
        <button class="primary" on:click={openPort} disabled={$busy}>Open Port</button>
        <button class="danger" on:click={closePort} disabled={$busy}>Close Port</button>
      </div>
    </div>

    <div class="hero-grid">
      <article class="stat-card">
        <span class="stat-label">Router</span>
        <strong class={$status?.discovered ? 'ok' : 'warn'}>
          {$status?.discovered ? 'Discovered' : 'Not selected'}
        </strong>
        <p>{ $status?.discovered ? 'Ready to manage mappings.' : 'Run discovery to pick a gateway.' }</p>
      </article>

      <article class="stat-card">
        <span class="stat-label">Mappings</span>
        <strong>{getPortCount()}</strong>
        <p>{getPortCount() === 0 ? 'No active mappings yet.' : 'Current mappings are listed below.'}</p>
      </article>

      <article class="stat-card warning">
        <span class="stat-label">Safety</span>
        <strong>Local only</strong>
        <p>Permanent exposure is off by default. Keep SSH/admin ports closed unless you really need them.</p>
      </article>
    </div>
  </section>

  <section class="layout">
    <div class="stack">
      <section class="panel">
        <div class="section-head">
          <div>
            <p class="eyebrow">Port mappings</p>
            <h2>Current state</h2>
          </div>
          <span class="badge">{getPortCount()} total</span>
        </div>

        {#if $status}
          {#if getPortCount() > 0}
            <div class="mapping-list">
              {#each $status.ports as port, index}
                <article class="mapping-item">
                  <div class="mapping-top">
                    <strong>Mapping {index + 1}</strong>
                    <span class="badge ok">Active</span>
                  </div>
                  <pre>{JSON.stringify(port, null, 2)}</pre>
                </article>
              {/each}
            </div>
          {:else}
            <div class="empty-state">
              <strong>No mappings yet</strong>
              <p>Open a port only when you need external access. For admin services, prefer closing them again after use.</p>
            </div>
          {/if}
        {:else}
          <div class="empty-state">
            <strong>Loading status…</strong>
            <p>Fetching router state and settings from the local backend.</p>
          </div>
        {/if}
      </section>

      <section class="panel">
        <div class="section-head">
          <div>
            <p class="eyebrow">Settings</p>
            <h2>Local config</h2>
          </div>
          <span class="badge">Saved with config.json</span>
        </div>

        <div class="field-grid">
          <label>
            <span>Listen address</span>
            <input bind:value={form.listen_addr} placeholder="127.0.0.1:8080" />
            <small>Keep this on localhost so the UI stays private.</small>
          </label>

          <label class="checkline">
            <input type="checkbox" bind:checked={form.auto_discover} />
            <div>
              <span>Auto discover on startup</span>
              <small>Turn this off only if you want to pick routers manually.</small>
            </div>
          </label>
        </div>

        <div class="actions actions-bottom">
          <button class="primary" on:click={save} disabled={$busy}>Save settings</button>
        </div>
      </section>
    </div>

    <aside class="stack">
      <section class="panel">
        <div class="section-head">
          <div>
            <p class="eyebrow">What to do next</p>
            <h2>Short guide</h2>
          </div>
        </div>

        <ol class="guide-list">
          <li>Reload the current status.</li>
          <li>Re-scan the router if nothing is discovered.</li>
          <li>Open a port only when you need it.</li>
          <li>Close risky mappings again when you're done.</li>
        </ol>
      </section>

      <section class="panel note">
        <p class="eyebrow">Security note</p>
        <p>SSH, admin, and database ports are the ones most likely to bite you later. Keep them off the internet unless there is a very specific reason to expose them.</p>
      </section>
    </aside>
  </section>
</main>
