<script lang="ts">

    import {getCurrentAppliedColorTheme, type MichuTechColorTheme, registerThemeChangeHandler} from '$lib/colors/ThemeLoader';
    import {onMount} from 'svelte';

    export let username: string;
    let url = ""

    onMount(() => {
        url = buildUrl(username, getCurrentAppliedColorTheme())
        registerThemeChangeHandler(t => url = buildUrl(username, t))
    })

    function buildUrl(name: string, theme: MichuTechColorTheme): string {
        let foregroundColor = theme === "dark" ? "D5D7D5" : "140D03"
        let backgroundColor = theme === "dark" ? "2b2826" : "d9d9d7"
        return `https://ui-avatars.com/api/?name=${name}&background=${backgroundColor}&color=${foregroundColor}`
    }
</script>

<img alt="dynamic profile" src={url} />