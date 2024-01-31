<script>
    import PageEditor from './PageEditor.svelte';
    import PageList from './PageList.svelte';
    import UserEditor from './UserEditor.svelte';
    import UserList from './UserList.svelte';

	export let apiUrl;

	let admin = false;
	let username;

	isLoggedIn()
	.then(user => {
        if(!user.admin)
            location.href = "/login?origin=/admin"

		admin = user.admin;
		username = user.mail;
	}).catch(_ => location.href = "/login?origin=/admin");

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
	<title>Micha Schweizer @ Admin</title>
</svelte:head>
<main>
	<h1>Micha Schweizer @ Admin</h1>
	{#if admin }
		<p class="page-description">
			Welcome {username}!
			<br />
			Here you can manage the pages and the admins of my deployments.
		</p>

		<h2>Pages</h2>
        <PageList apiUrl={apiUrl} />
        <PageEditor apiUrl={apiUrl} />

        <h2>Users</h2>
        <UserList apiUrl={apiUrl} />
        <UserEditor apiUrl={apiUrl} />
	{/if}
</main>