<script>
	import {user} from "./store.js";
    import {onMount} from 'svelte';

	export let apiUrl;
	let darkTheme = localStorage.getItem(THEME_STORAGE_KEY) === DARK_THEME;
    let pages = [];

    onMount(async () => {
        const loggedInUser = await isLoggedIn();
        if(loggedInUser) {
            user.set(loggedInUser);
            pages = loggedInUser.pages;
        } else {
            pages = await getPages();
        }
    });


	async function getPages() {
		const response = await fetch(`${apiUrl}/pages`);

		const stringResponse = await response.text();
		if(stringResponse.startsWith("]"))
			return [];
		return JSON.parse(stringResponse);
	}

	async function isLoggedIn() {
		const response = await fetch(`${apiUrl}/security/auth`);

		if (response.ok) {
			return (await response.json());
		} else {
			if (response.status === 401)
				return undefined;
			throw new Error("failed to check login status: " + response.status);
		}
	}
</script>

<svelte:head>
	<title>Micha Schweizer @ Home</title>
</svelte:head>
<section class="logo-banner">
	<svg class="icon" viewBox="0 0 374 374" fill="none" xmlns="http://www.w3.org/2000/svg">
		<g clip-path="url(#clip0_1_7)">
			<path id="michu-tech-logo-upper" fill="#4C4F4C" d="M0 59V0L188 186.5L298.5 74V299.5L334 336.5V42H75.5L33.5 0H374.5V374.5H314.5L261 320.5V166.5L184.5 244L0 59Z"/>
			<path id="michu-tech-logo-lower" fill="#E28521" d="M0 374.5V100.5L198 299H157.5L36 178V335.5H236L274 374.5H0Z"/>
		</g>
		<defs>
			<clipPath id="clip0_1_7">
				<rect width="512" height="512" fill="transparent"/>
			</clipPath>
		</defs>
	</svg>
	<p>michu - tech</p>
</section>
<main>
	<h1>Micha Schweizer @ Home</h1>
	<p class="page-description">
		This page shows all applications that are deployed on my server. Feel free to explore and
		get to know my projects more closely.
		<br />
		<span class="small-text">Some might me locked, you need to have special access for these.</span>
	</p>

    <div class="pages">
        {#each pages as page}
            <a class="page" class:disabled-page={page.privatePage && (!$user || !page.hasAccess)} href={page.url}>
                <h3 class="page-title">
                    {page.title}
                    <span class="small-text">{page.url}</span>
                </h3>
                <p class="text" style="">{page.description}</p>
                <span style="display: none">{page.privatePage}</span>

                {#if page.privatePage}
                    {#if $user && page.hasAccess}
                        <span class="material-symbols-rounded lock">lock_open</span>
                    {:else }
                        <span style="color: var(--michu-tech-warn);" class="material-symbols-rounded lock">encrypted</span>
                    {/if}
                {/if}
            </a>
        {/each}
    </div>
</main>
<div class="options">
	<button on:click={() => darkTheme = toggleDarkTheme()}>
		{#if darkTheme}
			<span class="material-symbols-rounded">light_mode</span>
		{:else }
			<span class="material-symbols-rounded">dark_mode</span>
		{/if}
	</button>
    {#if $user}
        <button on:click={() => darkTheme = toggleDarkTheme()}>
            <span class="material-symbols-rounded">settings</span>
        </button>
    {/if}
</div>

<style>
	.logo-banner {
		display: flex;
		flex-flow: row nowrap;
		gap: 8px;
		align-items: center;
	}

	.logo-banner .icon {
		width: auto;
		height: 42px;
		border-radius: 2px;
	}

	.logo-banner p {
		font-size: 38px;
		letter-spacing: -3px;
		margin: 0;
		text-transform: uppercase;
		font-family: 'Jura', sans-serif;
		font-weight: 700;
		color: var(--michu-tech-primary);
	}

	.pages {
		display: flex;
		flex-flow: row wrap;
		gap: 20px;
        justify-content: center;
	}

	.pages .page {
		all: unset;
		position: relative;

		padding: 20px;
		flex: 1;
        min-width: 220px;
        max-width: 600px;
        border-bottom-color: var(--michu-tech-foreground);
        border-bottom-style: dashed;
        border-bottom-width: 2px;

		cursor: pointer;
		transition: all 200ms ease-out;

        box-sizing: border-box;
	}

    .page-title,
    .page-title span {
        transition: all 200ms ease-out;
    }

	.disabled-page {
		pointer-events: none !important;
	}

	.pages .page:hover {
        border-bottom-color: var(--michu-tech-accent);
        border-bottom-style: dashed;
        border-bottom-width: 4px;
        scale: 1.02;
	}

    .pages .page:hover .page-title,
    .pages .page:hover .page-title span {
        color: var(--michu-tech-accent);
    }

	.lock {
		position: absolute;
		top: 10px;
		right: 10px;

		font-size: 40px;
	}

	.options {
		display: flex;
        flex-flow: row-reverse nowrap;

		position: fixed;
		right: 5vw;
		bottom: 5vh;
		width: auto;
		height: 60px;
	}

	.options button {
		all: unset;
		cursor: pointer;
		width: 55px;
		height: 55px;
		display: flex;
		justify-content: center;
		align-items: center;
		transition: all 250ms ease-out;
	}

    .options button span {
        transition: color 250ms ease-out;
    }

	.options button:hover {
		scale: 1.1;
		filter: brightness(110%);
	}

    .options button:hover span {
        color: var(--michu-tech-accent);
    }

	.options button .material-symbols-rounded {
		color: var(--michu-tech-foreground);
		font-size: 42px;
	}
</style>