<script>
	import {user} from "./store.js";
	import Login from "./Login.svelte";

	export let apiUrl;

	isLoggedIn()
	.then(loggedInUser => user.set(loggedInUser))
	.catch(e => console.log(e));

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
	<title>Overview</title>
</svelte:head>
<main>
	<h1>Overview</h1>
	{#await getPages()}
		<p>Loading pages ...</p>
	{:then pages}
		{#each pages as page}
			<p>{page.title}</p>
		{/each}
	{:catch error}
		<p>Could not load pages: {error.message}</p>
	{/await}

	{#if $user }
		<p>{$user.mail}</p>
	{:else }
		<div>
			<h3>Login to see more pages</h3>
			<Login bind:apiUrl/>
		</div>
	{/if}
</main>

<style>

</style>