<script>
    import {onMount} from 'svelte';
    import ChangePassword from '../../public/ChangePassword.svelte';

	export let apiUrl;

    let mail;
    let pageAccess = [];
	let errorText = "";

    onMount(() => {
        const urlSearchParams = new URLSearchParams(window.location.search);
        mail = urlSearchParams.get("mail");

        pageAccess = urlSearchParams.get("pages")?.split(",") ?? [];
    });

	async function sendActivateUser(data) {
		const response = await fetch(`${apiUrl}/security/activate`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ mail: data.mail, oldPassword: data.oldPassword, password: data.password }),
			credentials: 'include' // Include cookies in subsequent requests
		});

		if(response.ok) {
            errorText = "";
            location.href = '/';
        } else {
            errorText = "Failed to activate user. Are you sure that this is a different password or is your user already active?";
        }
	}
</script>

<svelte:head>
	<title>Micha Schweizer @ Onboarding</title>
</svelte:head>
<main>
	<h1>Micha Schweizer @ Onboarding</h1>
	<p class="page-description">
        Hi {mail ?? ""} <br/>
        You have been invited to access the following pages on my server. Please login and change your password to complete the setup.<br/>
        {#if pageAccess?.length > 0}
            <span class="small-text">Pages: {pageAccess.join(", ")}</span>
        {/if}
    </p>

    <ChangePassword saveText="Activate User" mail={mail} showCancel={false} showTitle={false} bind:errorText={errorText} on:save={(d) => sendActivateUser(d.detail)}/>
</main>

<style>
</style>