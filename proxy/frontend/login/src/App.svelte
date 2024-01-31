<script>
	import { fly } from 'svelte/transition';

	export let apiUrl;

	let mail;
	let password;
	$: valid = !mail || !password;
    let origin = buildNextUrl();

	let loginFailed = false;

	async function sendLogin() {
		const response = await fetch(`${apiUrl}/security/login`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ mail: mail, password: password }),
			credentials: 'include' // Include cookies in subsequent requests
		});

		loginFailed = !response.ok;
		if(response.ok)
			location.href = origin;
	}

    function buildNextUrl() {
        const props = new URLSearchParams(window.location.search);
        const origin = props.get("origin");
        return origin ?? "/";
    }
</script>

<svelte:head>
	<title>Micha Schweizer @ Login</title>
</svelte:head>
<main>
	<h1>Micha Schweizer @ Login</h1>
	<p class="page-description">Login here to access private pages and manage the deployments.</p>
	<form class="login-form">
		<input type="text" class="input" placeholder="Mail" bind:value={mail} autocomplete="username"/>
		<input type="password" class="input" placeholder="Password" bind:value={password} autocomplete="current-password"/>
		<div></div>
		<button type="submit" class="button" on:click|preventDefault={sendLogin} disabled={valid}>Login</button>
	</form>
	{#if loginFailed }
		<p transition:fly="{{delay: 0, duration: 300, y: -20 }}">Failed to login!</p>
	{/if}
</main>

<style>
	.login-form {
		display: flex;
		flex-flow: column;
        gap: 10px;

        align-items: center;
        justify-content: center;

        width: 100%;
		max-width: 84vw;
        height: 100%;

        padding-top: 20px;
	}

    input, button {
        min-width: 220px;
        width: 80vw;
        max-width: 360px;
    }
</style>