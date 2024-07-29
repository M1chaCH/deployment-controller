<script lang="ts">
    import {getCurrentAppliedColorTheme, initColorTheme, registerThemeChangeHandler, toggleDarkTheme} from '$lib/colors/ThemeLoader';
    import {onMount} from 'svelte';

    let darkThemeActive = getCurrentAppliedColorTheme() === "dark"

    onMount(() => {
        registerThemeChangeHandler(t => darkThemeActive = t === "dark")
        initColorTheme();
    })
</script>

<button on:click={() => toggleDarkTheme()} class="icon-button">
    {#if darkThemeActive}
        <span class="material-symbols-outlined">light_mode</span>
    {:else }
        <span class="material-symbols-outlined">dark_mode</span>
    {/if}
</button>
