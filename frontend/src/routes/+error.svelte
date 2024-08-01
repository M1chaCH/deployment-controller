<script lang="ts">
    import {page} from '$app/stores';
    import FetchFailed from '$lib/components/FetchFailed.svelte';
    import {onMount} from 'svelte';

    let statusText = "Page Error";
    onMount(() => {
        page.subscribe(p => {
            switch (p.status) {
                case 400:
                    statusText = "Bad Request"
                    return;
                case 404:
                    statusText = "Not Found"
                    return;
                case 401:
                    statusText = "Unauthorized"
                    return;
                case 403:
                    statusText = "Forbidden"
                    return;
                case 500:
                    statusText = "Server Error"
                    return;
            }
        })
    })
</script>

<div style="margin: 3rem;">
    <FetchFailed error={{
    message: $page.error?.message ?? "unknown",
    status: $page.status,
    statusText: statusText,
}} />
</div>