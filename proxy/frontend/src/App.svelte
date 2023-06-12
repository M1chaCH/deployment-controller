<script>
	import {user} from "./store.js";
	import Login from "./Login.svelte";

	export let apiUrl;

	isLoggedIn()
	.then(loggedInUser => user.set(loggedInUser))
	.catch(_ => console.log("failed to check login status"));

	let viewPrivate = false;
	user.subscribe(u => viewPrivate = u?.viewPrivate || false);

	async function getPages() {
		const response = await fetch(`${apiUrl}/pages`);
		return await response.json();
	}

	async function isLoggedIn() {
		const response = await fetch(`${apiUrl}/security/auth`);

		if(response.ok) {
			return (await response.json());
		} else {
			if(response.status === 401)
				return undefined;
			throw new Error("failed to check login status: " + response.status);
		}
	}
</script>

<svelte:head>
	<title>Micha Schweizer @ Home</title>
</svelte:head>
<main>
	<h1>Micha Schweizer @ Home</h1>
	<p style="max-width: 500px; font-size: 20px; font-weight: 300; padding-bottom: 5vh;">
		This page shows all applications that are deployed on my server. Feel free to explore and
		get to know my projects more closely.
		<br />
		<span class="small-text">Some might me locked, you need to have special access for these.</span>
	</p>
	{#await getPages()}
		<p>Loading pages ...</p>
	{:then pages}
		<div class="pages">
			{#each pages as page}
				<a class="page" class:disabled-page={!viewPrivate && page.private_access} href={page.url}>
					<h3>
						{page.title}
						<span class="small-text">{page.url}</span>
					</h3>
					<p class="text" style="">{page.description}</p>
					<span style="display: none">{page.private_access}</span>

					{#if page.private_access}
						{#if viewPrivate}
							<span class="material-symbols-rounded lock">lock_open</span>
						{:else }
							<span class="material-symbols-rounded lock">encrypted</span>
						{/if}
					{/if}
				</a>
			{/each}
		</div>
	{:catch error}
		<p>Could not load pages: {error.message}</p>
	{/await}

	{#if $user }
		<p>{$user.mail}</p>
	{:else }
		<div>
<!--			<h3>Login to see more pages</h3>-->
<!--			<Login bind:apiUrl/>-->
		</div>
	{/if}
</main>

<style>
	.pages {
		display: flex;
		flex-flow: row wrap;
		gap: 20px;
	}

	.pages .page {
		all: unset;
		position: relative;

		padding: 20px;
		width: 350px;
		background-color: #2A9D8F;
		box-shadow: inset 0 0 0 4px #264653;

		cursor: pointer;
		transition: all 250ms ease-out;
	}

	.disabled-page {
		pointer-events: none !important;
	    background-color: color-mix(in srgb, #2A9D8F 25%, transparent) !important;
	    box-shadow: inset 0 0 0 4px color-mix(in srgb, #264653 60%, white) !important;
	}

	.pages .page:hover {
		transition: all 250ms ease-out;
		background-color: #264653;

		box-shadow: inset -10px -10px 0 10px #2A9D8F;
	}

	.pages .page:hover h3,
	.pages .page:hover span,
	.pages .page:hover p {
		color: white;
	}

	.lock {
		position: absolute;
		top: 10px;
		right: 10px;

		font-size: 40px;
	}
</style>