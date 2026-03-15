<script>
    import {onMount} from 'svelte'
    import {fade, fly} from 'svelte/transition'
    import {
        CreateEntry,
        CreateVault,
        DeleteEntry,
        GeneratePassword,
        GetTOTPCode,
        HasVault,
        ListEntries,
        ListVaults,
        Lock,
        Unlock,
        UpdateEntry
    } from '../wailsjs/go/gui/App.js'

    let screen = 'loading'
    let vaults = []
    let entries = []
    let selectedVault = ''
    let password = ''
    let vaultName = ''
    let error = ''
    let search = ''
    let selected = null   // entry open in detail/form
    let isEditing = false
    let showPassword = false
    let showFormPassword = false
    let copied = ''
    let form = emptyForm()
    let detailPanel = null

    function resetDetailScroll() {
        if (detailPanel) detailPanel.scrollTop = 0
    }

    // TOTP
    let totpState = null   // { code, remaining }
    let totpTimer = null

    function emptyForm() {
        return {id: null, title: '', username: '', password: '', url: '', notes: '', totpSecret: '', totpIssuer: ''}
    }

    async function refreshTOTP(id) {
        if (!id) {
            totpState = null;
            return
        }
        try {
            totpState = await GetTOTPCode(id)
        } catch {
            totpState = null
        }
    }

    function startTOTPTimer(entry) {
        stopTOTPTimer()
        if (!entry?.totpSecret) return
        refreshTOTP(entry.id)
        totpTimer = setInterval(() => refreshTOTP(entry.id), 1000)
    }

    function stopTOTPTimer() {
        if (totpTimer) {
            clearInterval(totpTimer);
            totpTimer = null
        }
        totpState = null
    }

    onMount(async () => {
        const has = await HasVault()
        if (has) {
            vaults = await ListVaults()
            selectedVault = vaults[0] ?? ''
            screen = 'unlock'
        } else {
            screen = 'create'
        }
    })

    async function handleCreate() {
        error = ''
        try {
            await CreateVault(vaultName, password)
            vaults = await ListVaults()
            selectedVault = vaultName
            await loadEntries()
        } catch (e) {
            error = String(e)
        }
    }

    async function handleUnlock() {
        error = ''
        try {
            await Unlock(selectedVault, password)
            await loadEntries()
        } catch (e) {
            error = 'Invalid master password'
        }
    }

    async function loadEntries() {
        entries = await ListEntries()
        screen = 'app'
        password = ''
        selected = null
    }

    async function handleLock() {
        stopTOTPTimer()
        await Lock()
        screen = 'unlock'
        password = ''
        entries = []
        selected = null
    }

    function selectEntry(entry) {
        selected = entry
        isEditing = false
        showPassword = false
        copied = ''
        startTOTPTimer(entry)
        resetDetailScroll()
    }

    function openNew() {
        stopTOTPTimer()
        form = emptyForm()
        selected = null
        isEditing = true
        showFormPassword = false
        resetDetailScroll()
    }

    function startEdit() {
        stopTOTPTimer()
        form = {
            id: selected.id,
            title: selected.title ?? '',
            username: selected.username ?? '',
            password: selected.password ?? '',
            url: selected.url ?? '',
            notes: selected.notes ?? '',
            totpSecret: selected.totpSecret ?? '',
            totpIssuer: selected.totpIssuer ?? '',
        }
        isEditing = true
        showFormPassword = false
        resetDetailScroll()
    }

    async function handleSave() {
        error = ''
        try {
            const dto = {
                id: form.id,
                title: form.title,
                username: form.username || null,
                password: form.password || null,
                url: form.url || null,
                notes: form.notes || null,
                totpSecret: form.totpSecret || null,
                totpIssuer: form.totpIssuer || null,
                totpDigits: 6,
                totpPeriod: 30,
                totpAlgorithm: 'SHA1',
            }
            if (form.id) {
                await UpdateEntry(form.id, dto)
            } else {
                await CreateEntry(dto)
            }
            await loadEntries()
            const updated = entries.find(e => e.title === form.title)
            if (updated) selectEntry(updated)
        } catch (e) {
            error = String(e)
        }
    }

    async function handleDelete() {
        if (!selected) return
        await DeleteEntry(selected.id)
        selected = null
        await loadEntries()
    }

    async function genPassword() {
        form.password = await GeneratePassword(20)
    }

    async function copyText(text, key) {
        await navigator.clipboard.writeText(text)
        copied = key
        setTimeout(() => copied = '', 2000)
    }

    function initials(title) {
        return (title ?? '?').slice(0, 2).toUpperCase()
    }

    function avatarColor(title) {
        const colors = ['#7c3aed', '#2563eb', '#0891b2', '#059669', '#d97706', '#dc2626', '#db2777']
        let h = 0
        for (let c of title ?? '') h = (h * 31 + c.charCodeAt(0)) & 0xffff
        return colors[h % colors.length]
    }

    $: filtered = entries.filter(e =>
        (e.title ?? '').toLowerCase().includes(search.toLowerCase()) ||
        (e.username ?? '').toLowerCase().includes(search.toLowerCase())
    )
</script>

<!-- ───────── AUTH SCREENS ───────── -->
{#if screen === 'loading'}
    <div class="auth-bg">
        <div class="spinner"></div>
    </div>

{:else if screen === 'create' || screen === 'unlock'}
    <div class="auth-bg" transition:fade>
        <div class="auth-card" in:fly="{{ y: 20, duration: 300 }}">
            <div class="auth-logo">
                <svg width="48" height="48" viewBox="0 0 40 40" fill="none">
                    <rect width="40" height="40" rx="10" fill="#7c3aed"/>
                    <path d="M20 8a8 8 0 0 0-8 8v2h-2v14h20V18h-2v-2a8 8 0 0 0-8-8zm0 3a5 5 0 0 1 5 5v2H15v-2a5 5 0 0 1 5-5zm0 10a3 3 0 1 1 0 6 3 3 0 0 1 0-6z"
                          fill="white"/>
                </svg>
                <span class="auth-app-name">KYP</span>
                <span class="auth-app-sub">Keep your passwords</span>
            </div>

            <h2>{screen === 'create' ? 'Create your vault' : 'Welcome back'}</h2>
            <p class="auth-sub">{screen === 'create' ? 'Set a strong master password to protect your data.' : 'Enter your master password to unlock.'}</p>

            {#if error}
                <div class="auth-error">{error}</div>
            {/if}

            {#if screen === 'create'}
                <div class="field">
                    <label for="vault-name">Vault name</label>
                    <input name="vault-name" bind:value={vaultName} placeholder="My Vault"/>
                </div>
            {:else if vaults.length > 1}
                <div class="field">
                    <label for="vault">Vault</label>
                    <select name="vault" bind:value={selectedVault}>
                        {#each vaults as v}
                            <option>{v}</option>
                        {/each}
                    </select>
                </div>
            {:else}
                <div class="vault-badge">{selectedVault}</div>
            {/if}

            <div class="field">
                <label for="master-password">Master password</label>
                <input
                        name="master-password"
                        type="password"
                        bind:value={password}
                        placeholder="••••••••••••"
                        on:keydown={e => e.key === 'Enter' && (screen === 'create' ? handleCreate() : handleUnlock())}
                />
            </div>

            <button class="btn-primary full" on:click={screen === 'create' ? handleCreate : handleUnlock}>
                {screen === 'create' ? 'Create vault' : 'Unlock'}
            </button>
        </div>
    </div>

{:else if screen === 'app'}
    <div class="app" transition:fade>
        <aside class="sidebar">
            <div class="sidebar-header">
                <div class="sidebar-logo">
                    <svg width="24" height="24" viewBox="0 0 40 40" fill="none">
                        <rect width="40" height="40" rx="10" fill="#7c3aed"/>
                        <path d="M20 8a8 8 0 0 0-8 8v2h-2v14h20V18h-2v-2a8 8 0 0 0-8-8zm0 3a5 5 0 0 1 5 5v2H15v-2a5 5 0 0 1 5-5zm0 10a3 3 0 1 1 0 6 3 3 0 0 1 0-6z"
                              fill="white"/>
                    </svg>
                    <span>kyp</span>
                </div>
                <button class="icon-btn" title="Lock vault" on:click={handleLock}>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="11" width="18" height="11" rx="2"/>
                        <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                    </svg>
                </button>
            </div>

            <div class="sidebar-search">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="11" cy="11" r="8"/>
                    <path d="m21 21-4.35-4.35"/>
                </svg>
                <input bind:value={search} placeholder="Search…"/>
            </div>

            <div class="sidebar-actions">
                <button class="new-btn" on:click={openNew}>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                         stroke-width="2.5">
                        <path d="M12 5v14M5 12h14"/>
                    </svg>
                    New entry
                </button>
            </div>

            <nav class="entry-list">
                {#each filtered as entry (entry.id)}
                    <button
                            class="entry-row"
                            class:active={selected?.id === entry.id}
                            on:click={() => selectEntry(entry)}
                    >
                        <div class="avatar" style="background:{avatarColor(entry.title)}">
                            {initials(entry.title)}
                        </div>
                        <div class="entry-info">
                            <span class="entry-name">{entry.title}</span>
                            <span class="entry-user">{entry.username ?? '—'}</span>
                        </div>
                    </button>
                {:else}
                    <p class="empty-list">{search ? 'No results' : 'No entries yet'}</p>
                {/each}
            </nav>
        </aside>

        <main class="detail" bind:this={detailPanel}>
            {#if isEditing}
                <div class="detail-inner" in:fly="{{ x: 10, duration: 200 }}">
                    <div class="detail-header">
                        <h2>{form.id ? 'Edit entry' : 'New entry'}</h2>
                        <button class="icon-btn" on:click={() => { isEditing = false; if (!form.id) selected = null; resetDetailScroll() }}>
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                 stroke-width="2">
                                <path d="M18 6 6 18M6 6l12 12"/>
                            </svg>
                        </button>
                    </div>

                    {#if error}
                        <div class="form-error">{error}</div>
                    {/if}

                    <div class="form-fields">
                        <div class="field">
                            <label for="title">Title <span class="req">*</span></label>
                            <input name="title" bind:value={form.title} placeholder="e.g. GitHub"/>
                        </div>
                        <div class="field">
                            <label for="username">Username / Email</label>
                            <input name="username" bind:value={form.username} placeholder="user@example.com"/>
                        </div>
                        <div class="field">
                            <label for="password">Password</label>
                            <div class="input-row">
                                {#if showFormPassword}
                                    <input name="password" bind:value={form.password} placeholder="password"/>
                                {:else}
                                    <input name="password" type="password" bind:value={form.password}
                                           placeholder="••••••••"/>
                                {/if}
                                <button class="icon-btn-sm" title="Show/hide"
                                        on:click={() => showFormPassword = !showFormPassword}>
                                    {#if showFormPassword}
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                             stroke="currentColor" stroke-width="2">
                                            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                                            <line x1="1" y1="1" x2="23" y2="23"/>
                                        </svg>
                                    {:else}
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                             stroke="currentColor" stroke-width="2">
                                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                                            <circle cx="12" cy="12" r="3"/>
                                        </svg>
                                    {/if}
                                </button>
                                <button class="icon-btn-sm generate-btn" title="Generate password"
                                        on:click={genPassword}>
                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                         stroke-width="2">
                                        <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
                                    </svg>
                                </button>
                            </div>
                        </div>
                        <div class="field">
                            <label for="url">Website URL</label>
                            <input name="url" bind:value={form.url} placeholder="https://…"/>
                        </div>
                        <div class="field">
                            <label for="notes">Notes</label>
                            <textarea name="notes" bind:value={form.notes} rows="4"
                                      placeholder="Any additional notes…"></textarea>
                        </div>
                        <div class="totp-section">
                            <div class="totp-section-label">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                     stroke-width="2">
                                    <circle cx="12" cy="12" r="10"/>
                                    <polyline points="12 6 12 12 16 14"/>
                                </svg>
                                Two-Factor Authentication (TOTP)
                            </div>
                            <div class="field">
                                <label for="totp-secret">Secret key</label>
                                <input id="totp-secret" bind:value={form.totpSecret} placeholder="JBSWY3DPEHPK3PXP"
                                       autocomplete="off"/>
                            </div>
                            <div class="field">
                                <label for="totp-issuer">Issuer (optional)</label>
                                <input id="totp-issuer" bind:value={form.totpIssuer} placeholder="e.g. GitHub"/>
                            </div>
                        </div>
                    </div>

                    <div class="form-actions">
                        <button class="btn-primary" on:click={handleSave}>Save</button>
                        <button class="btn-ghost" on:click={() => { isEditing = false; if (!form.id) selected = null; resetDetailScroll() }}>
                            Cancel
                        </button>
                    </div>
                </div>

            {:else if selected}
                <div class="detail-inner" in:fly="{{ x: 10, duration: 200 }}">
                    <div class="detail-header">
                        <div class="detail-avatar" style="background:{avatarColor(selected.title)}">
                            {initials(selected.title)}
                        </div>
                        <div class="detail-title-block">
                            <h2>{selected.title}</h2>
                            {#if selected.url}<a href={selected.url} class="detail-url"
                                                 target="_blank">{selected.url}</a>{/if}
                        </div>
                        <div class="header-actions">
                            <button class="btn-ghost sm" on:click={startEdit}>Edit</button>
                            <button class="btn-danger sm" on:click={handleDelete}>Delete</button>
                        </div>
                    </div>

                    <div class="detail-fields">
                        {#if selected.username}
                            <div class="detail-field">
                                <span class="field-label">Username</span>
                                <div class="field-value-row">
                                    <span class="field-value">{selected.username}</span>
                                    <button class="icon-copy-btn" class:copied={copied === 'username'}
                                            on:click={() => copyText(selected.username, 'username')}>
                                        {#if copied === 'username'}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2.5">
                                                <polyline points="20 6 9 17 4 12"/>
                                            </svg>
                                        {:else}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2">
                                                <rect x="9" y="9" width="13" height="13" rx="2"/>
                                                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                                            </svg>
                                        {/if}
                                    </button>
                                </div>
                            </div>
                        {/if}

                        {#if selected.password}
                            <div class="detail-field">
                                <span class="field-label">Password</span>
                                <div class="field-value-row">
                                    <span class="field-value mono">{showPassword ? selected.password : '••••••••••••'}</span>
                                    <button class="icon-btn-sm" on:click={() => showPassword = !showPassword}
                                            title="Show/hide">
                                        {#if showPassword}
                                            <svg width="15" height="15" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2">
                                                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                                                <line x1="1" y1="1" x2="23" y2="23"/>
                                            </svg>
                                        {:else}
                                            <svg width="15" height="15" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2">
                                                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                                                <circle cx="12" cy="12" r="3"/>
                                            </svg>
                                        {/if}
                                    </button>
                                    <button class="icon-copy-btn" class:copied={copied === 'password'}
                                            on:click={() => copyText(selected.password, 'password')}>
                                        {#if copied === 'password'}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2.5">
                                                <polyline points="20 6 9 17 4 12"/>
                                            </svg>
                                        {:else}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2">
                                                <rect x="9" y="9" width="13" height="13" rx="2"/>
                                                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                                            </svg>
                                        {/if}
                                    </button>
                                </div>
                            </div>
                        {/if}

                        {#if selected.url}
                            <div class="detail-field">
                                <span class="field-label">URL</span>
                                <div class="field-value-row">
                                    <a href={selected.url} target="_blank" class="field-value link">{selected.url}</a>
                                    <button class="icon-copy-btn" class:copied={copied === 'url'}
                                            on:click={() => copyText(selected.url, 'url')}>
                                        {#if copied === 'url'}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2.5">
                                                <polyline points="20 6 9 17 4 12"/>
                                            </svg>
                                        {:else}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                 stroke="currentColor" stroke-width="2">
                                                <rect x="9" y="9" width="13" height="13" rx="2"/>
                                                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                                            </svg>
                                        {/if}
                                    </button>
                                </div>
                            </div>
                        {/if}

                        {#if selected.notes}
                            <div class="detail-field">
                                <span class="field-label">Notes</span>
                                <p class="field-notes">{selected.notes}</p>
                            </div>
                        {/if}

                        {#if selected.totpSecret && totpState}
                            {@const period = selected.totpPeriod || 30}
                            {@const half = Math.floor(totpState.code.length / 2)}
                            {@const
                                accent = totpState.remaining <= 5 ? '#ef4444' : totpState.remaining <= 10 ? '#f59e0b' : '#7c3aed'}
                            <div class="detail-field">
                                <span class="field-label">{selected.totpIssuer ? selected.totpIssuer + ' · ' : ''}2FA Code</span>
                                <div class="totp-widget">
                                    <div class="totp-code-wrap">
                                        {#each totpState.code.split('') as digit, i}
                                            {#if i === half}
                                                <span class="totp-sep"> </span>
                                            {/if}
                                            <div class="digit-roller">
                                                <div class="digit-strip"
                                                     style="transform: translateY(-{parseInt(digit) * 10}%)">
                                                    {#each [0, 1, 2, 3, 4, 5, 6, 7, 8, 9] as d}
                                                        <span>{d}</span>
                                                    {/each}
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                    <div class="totp-right">
                                        <svg class="totp-ring" viewBox="0 0 36 36" width="36" height="36">
                                            <circle cx="18" cy="18" r="15" fill="none" stroke="#e5e5ea"
                                                    stroke-width="3"/>
                                            <circle cx="18" cy="18" r="15" fill="none"
                                                    stroke={accent}
                                                    stroke-width="3"
                                                    stroke-dasharray="{(totpState.remaining / period) * 94.25} 94.25"
                                                    stroke-dashoffset="23.56"
                                                    stroke-linecap="round"
                                                    style="transition: stroke-dasharray .9s linear, stroke .3s"
                                            />
                                            <text x="18" y="22" text-anchor="middle" font-size="10"
                                                  fill="#666">{totpState.remaining}</text>
                                        </svg>
                                        <button class="icon-copy-btn" class:copied={copied === 'totp'} title="Copy code"
                                                on:click={() => copyText(totpState.code, 'totp')}>
                                            {#if copied === 'totp'}
                                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                     stroke="currentColor" stroke-width="2.5">
                                                    <polyline points="20 6 9 17 4 12"/>
                                                </svg>
                                            {:else}
                                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
                                                     stroke="currentColor" stroke-width="2">
                                                    <rect x="9" y="9" width="13" height="13" rx="2"/>
                                                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                                                </svg>
                                            {/if}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/if}
                    </div>
                </div>

            {:else}
                <div class="empty-detail">
                    <svg width="64" height="64" viewBox="0 0 40 40" fill="none" opacity=".15">
                        <rect width="40" height="40" rx="10" fill="#7c3aed"/>
                        <path d="M20 8a8 8 0 0 0-8 8v2h-2v14h20V18h-2v-2a8 8 0 0 0-8-8zm0 3a5 5 0 0 1 5 5v2H15v-2a5 5 0 0 1 5-5zm0 10a3 3 0 1 1 0 6 3 3 0 0 1 0-6z"
                              fill="#7c3aed"/>
                    </svg>
                    <p>Select an entry or create a new one</p>
                </div>
            {/if}
        </main>
    </div>
{/if}

<style>
    :global(*, *::before, *::after) {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }

    :global(body) {
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        background: #0f0f13;
        color: #1d1d1f;
        height: 100vh;
        overflow: hidden;
    }

    /* ── Auth ── */
    .auth-bg {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100vh;
        background: radial-gradient(ellipse at 60% 40%, #2d1b69 0%, #0f0f13 70%);
    }

    .spinner {
        width: 32px;
        height: 32px;
        border: 3px solid #7c3aed44;
        border-top-color: #7c3aed;
        border-radius: 50%;
        animation: spin .7s linear infinite;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    .auth-card {
        background: #1a1a24;
        border: 1px solid #2a2a3a;
        border-radius: 16px;
        padding: 2.5rem;
        width: 380px;
        box-shadow: 0 24px 64px rgba(0, 0, 0, .6);
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .auth-logo {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: .3rem;
        margin-bottom: 1.25rem;
    }

    .auth-app-name {
        font-size: 1.6rem;
        font-weight: 900;
        color: #fff;
        letter-spacing: .02em;
    }

    .auth-app-sub {
        font-size: .78rem;
        color: #666;
        letter-spacing: .04em;
        text-transform: uppercase;
    }

    .auth-card h2 {
        font-size: 1.2rem;
        font-weight: 700;
        color: #fff;
        text-align: center;
    }

    .auth-sub {
        font-size: .85rem;
        color: #888;
        line-height: 1.5;
    }

    .auth-error {
        background: #3d0f0f;
        border: 1px solid #7f1d1d;
        color: #fca5a5;
        padding: .6rem .9rem;
        border-radius: 8px;
        font-size: .85rem;
    }

    .vault-badge {
        background: #252535;
        border: 1px solid #333;
        border-radius: 8px;
        padding: .5rem .9rem;
        color: #ccc;
        font-size: .95rem;
    }

    .field {
        display: flex;
        flex-direction: column;
        gap: .35rem;
        text-align: left;
    }

    .field label {
        font-size: .8rem;
        color: #888;
        font-weight: 500;
        letter-spacing: .02em;
        text-transform: uppercase;
    }

    .field input, .field select, .field textarea {
        background: #252535;
        border: 1px solid #333;
        border-radius: 8px;
        padding: .6rem .9rem;
        color: #fff;
        font-size: .95rem;
        outline: none;
        width: 100%;
        transition: border-color .15s;
    }

    .field input:focus, .field select:focus, .field textarea:focus {
        border-color: #7c3aed;
    }

    .field textarea {
        resize: vertical;
        min-height: 80px;
    }

    .field select option {
        background: #1a1a24;
    }

    .btn-primary {
        background: #7c3aed;
        color: #fff;
        border: none;
        border-radius: 9px;
        padding: .65rem 1.2rem;
        font-size: .95rem;
        font-weight: 600;
        cursor: pointer;
        transition: background .15s, transform .1s;
    }

    .btn-primary:hover {
        background: #6d28d9;
    }

    .btn-primary:active {
        transform: scale(.98);
    }

    .btn-primary.full {
        width: 100%;
        padding: .75rem;
    }

    /* ── App shell ── */
    .app {
        display: flex;
        height: 100vh;
    }

    /* ── Sidebar ── */
    .sidebar {
        width: 280px;
        flex-shrink: 0;
        background: #13131c;
        border-right: 1px solid #1e1e2e;
        display: flex;
        flex-direction: column;
    }

    .sidebar-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 1rem 1rem .75rem;
        border-bottom: 1px solid #1e1e2e;
    }

    .sidebar-logo {
        display: flex;
        align-items: center;
        gap: .5rem;
    }

    .sidebar-logo span {
        color: #fff;
        font-weight: 800;
        font-size: 1.05rem;
    }

    .sidebar-search {
        display: flex;
        align-items: center;
        gap: .5rem;
        background: #1e1e2e;
        border-radius: 8px;
        padding: .45rem .75rem;
        margin: .75rem;
        border: 1px solid #2a2a3a;
    }

    .sidebar-search svg {
        color: #555;
        flex-shrink: 0;
    }

    .sidebar-search input {
        background: none;
        border: none;
        outline: none;
        color: #ddd;
        font-size: .9rem;
        width: 100%;
    }

    .sidebar-search input::placeholder {
        color: #555;
    }

    .sidebar-actions {
        padding: 0 .75rem .5rem;
    }

    .new-btn {
        display: flex;
        align-items: center;
        gap: .5rem;
        width: 100%;
        background: #7c3aed18;
        border: 1px solid #7c3aed44;
        color: #a78bfa;
        border-radius: 8px;
        padding: .5rem .75rem;
        font-size: .85rem;
        font-weight: 600;
        cursor: pointer;
        transition: background .15s;
    }

    .new-btn:hover {
        background: #7c3aed30;
    }

    .entry-list {
        flex: 1;
        overflow-y: auto;
        padding: .25rem .5rem;
    }

    .entry-list::-webkit-scrollbar {
        width: 4px;
    }

    .entry-list::-webkit-scrollbar-thumb {
        background: #2a2a3a;
        border-radius: 2px;
    }

    .entry-row {
        display: flex;
        align-items: center;
        gap: .75rem;
        width: 100%;
        background: none;
        border: none;
        border-radius: 8px;
        padding: .6rem .65rem;
        cursor: pointer;
        text-align: left;
        transition: background .12s;
    }

    .entry-row:hover {
        background: #1e1e2e;
    }

    .entry-row.active {
        background: #7c3aed22;
    }

    .avatar {
        width: 36px;
        height: 36px;
        border-radius: 9px;
        flex-shrink: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: .75rem;
        font-weight: 800;
        color: #fff;
        letter-spacing: .02em;
    }

    .entry-info {
        display: flex;
        flex-direction: column;
        min-width: 0;
    }

    .entry-name {
        color: #e2e2e8;
        font-size: .88rem;
        font-weight: 600;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .entry-user {
        color: #555;
        font-size: .78rem;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        margin-top: 1px;
    }

    .empty-list {
        color: #444;
        font-size: .85rem;
        text-align: center;
        padding: 2rem 1rem;
    }

    /* ── Detail / Form ── */
    .detail {
        flex: 1;
        background: #f8f8fc;
        overflow-y: auto;
        display: flex;
        flex-direction: column;
    }

    .detail-inner {
        padding: 2rem;
        max-width: 560px;
        width: 100%;
        margin: 0 auto;
    }

    .detail-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
        margin-bottom: 2rem;
        flex-wrap: wrap;
    }

    .detail-avatar {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        flex-shrink: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: .95rem;
        font-weight: 800;
        color: #fff;
    }

    .detail-title-block {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        flex: 1;
        min-width: 0;
    }

    .detail-header h2 {
        font-size: 1.3rem;
        font-weight: 700;
        color: #111;
        text-align: left;
    }

    .detail-url {
        font-size: .8rem;
        color: #7c3aed;
        text-decoration: none;
        display: block;
        margin-top: 2px;
    }

    .detail-url:hover {
        text-decoration: underline;
    }

    .header-actions {
        display: flex;
        gap: .5rem;
        margin-left: auto;
    }

    .detail-fields {
        display: flex;
        flex-direction: column;
        gap: 1px;
        border-radius: 12px;
        overflow: hidden;
        border: 1px solid #e5e5ea;
        background: #e5e5ea;
    }

    .detail-field {
        background: #fff;
        padding: .85rem 1.1rem;
        display: flex;
        flex-direction: column;
        gap: .3rem;
        align-items: stretch;
    }

    .field-label {
        font-size: .72rem;
        font-weight: 600;
        color: #999;
        text-transform: uppercase;
        letter-spacing: .06em;
        text-align: left;
    }

    .field-value-row {
        display: flex;
        align-items: center;
        gap: .5rem;
        width: 100%;
    }

    .field-value {
        font-size: .95rem;
        color: #111;
        flex: 1;
        word-break: break-all;
        text-align: left;
    }

    .field-value.mono {
        font-family: 'SF Mono', 'Fira Code', monospace;
        letter-spacing: .05em;
    }

    .field-notes {
        font-size: .9rem;
        color: #333;
        line-height: 1.6;
        white-space: pre-wrap;
    }

    /* TOTP widget */
    .totp-widget {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
        padding: .25rem 0;
    }

    .totp-code-wrap {
        display: flex;
        align-items: center;
        gap: .05em;
        font-family: 'SF Mono', 'Fira Code', monospace;
        font-size: 1.75rem;
        font-weight: 700;
        color: #7c3aed;
    }

    .digit-roller {
        display: inline-block;
        height: 1em;
        overflow: hidden;
        position: relative;
    }

    .digit-strip {
        display: flex;
        flex-direction: column;
        transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
        will-change: transform;
    }

    .digit-strip span {
        display: block;
        height: 1em;
        line-height: 1;
        text-align: center;
    }

    .totp-sep {
        width: .45em;
    }

    .totp-right {
        display: flex;
        align-items: center;
        gap: .75rem;
    }

    .totp-ring {
        display: block;
        flex-shrink: 0;
    }

    .icon-copy-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        border-radius: 8px;
        border: none;
        cursor: pointer;
        background: #f0f0f5;
        color: #666;
        transition: background .12s, color .12s;
    }

    .icon-copy-btn:hover {
        background: #e0e0ea;
        color: #333;
    }

    .icon-copy-btn.copied {
        background: #d1fae5;
        color: #065f46;
    }

    /* TOTP section in form */
    .totp-section {
        border-top: 1px dashed #e0e0ea;
        padding-top: 1rem;
        display: flex;
        flex-direction: column;
        gap: .85rem;
    }

    .totp-section-label {
        display: flex;
        align-items: center;
        gap: .4rem;
        font-size: .8rem;
        font-weight: 600;
        color: #888;
        text-transform: uppercase;
        letter-spacing: .04em;
    }

    .link {
        color: #7c3aed;
        text-decoration: none;
    }

    .link:hover {
        text-decoration: underline;
    }


    .empty-detail {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        color: #999;
        font-size: .9rem;
    }

    /* Form */
    .form-fields {
        display: flex;
        flex-direction: column;
        gap: 1.1rem;
        margin-bottom: 1.5rem;
    }

    .req {
        color: #e11d48;
    }

    .input-row {
        display: flex;
        gap: .4rem;
    }

    .input-row input {
        flex: 1;
    }

    .form-error {
        background: #fff1f2;
        border: 1px solid #fecdd3;
        color: #be123c;
        padding: .6rem .9rem;
        border-radius: 8px;
        font-size: .85rem;
        margin-bottom: 1rem;
    }

    .form-actions {
        display: flex;
        gap: .75rem;
    }

    .btn-ghost {
        background: transparent;
        border: 1px solid #ddd;
        color: #555;
        border-radius: 9px;
        padding: .65rem 1.2rem;
        font-size: .95rem;
        font-weight: 600;
        cursor: pointer;
        transition: border-color .15s, color .15s;
    }

    .btn-ghost:hover {
        border-color: #999;
        color: #222;
    }

    .btn-ghost.sm {
        padding: .4rem .85rem;
        font-size: .85rem;
    }

    .btn-danger {
        background: #fff1f2;
        border: 1px solid #fecdd3;
        color: #be123c;
        border-radius: 9px;
        padding: .4rem .85rem;
        font-size: .85rem;
        font-weight: 600;
        cursor: pointer;
        transition: background .15s;
    }

    .btn-danger:hover {
        background: #ffe4e6;
    }

    .icon-btn {
        background: none;
        border: none;
        color: #666;
        cursor: pointer;
        padding: .35rem;
        border-radius: 6px;
        display: flex;
        align-items: center;
        transition: background .12s, color .12s;
    }

    .icon-btn:hover {
        background: #1e1e2e;
        color: #aaa;
    }

    .icon-btn-sm {
        background: #f0f0f5;
        border: none;
        color: #666;
        cursor: pointer;
        padding: .4rem;
        border-radius: 6px;
        display: flex;
        align-items: center;
        transition: background .12s;
    }

    .icon-btn-sm:hover {
        background: #e0e0ea;
    }

    .generate-btn {
        color: #7c3aed;
    }

    .generate-btn:hover {
        background: #ede9fe;
    }

    /* Auth field overrides (dark theme) */
    .auth-card .field input,
    .auth-card .field select {
        background: #252535;
        border: 1px solid #333;
        color: #fff;
    }

    .auth-card .field input:focus,
    .auth-card .field select:focus {
        border-color: #7c3aed;
    }
</style>
