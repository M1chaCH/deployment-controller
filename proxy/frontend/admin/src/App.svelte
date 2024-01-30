<script>
	import { fly } from 'svelte/transition';

	export let apiUrl;

	let admin = true;
	let username;

	let pageIdToDelete = undefined;
	let pageToCreateId = "";
	let pageToCreateTitle = "";
	let pageToCreateDesc = "";
	let pageToCreateUrl = "";
	let pageToCreatePrivate = true;
	$: pageToCreateInvalid = !pageToCreateTitle || !pageToCreateDesc || !pageToCreateUrl || !pageToCreateUrl.startsWith("/");

	let userIdToDelete = undefined;
	let userToCreateMail = "";
	let userToCreatePassword = "";
	let userToCreateAdmin = false;
	let userToCreatePrivate = false;
	$: userToCreateInvalid = !userToCreateMail || !userToCreatePassword;

	isLoggedIn()
	.then(user => {
		admin = user.admin;
		username = user.mail;
	}).catch(_ => location.href = "/login");

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

	async function getUsers() {
		const response = await fetch(`${apiUrl}/users`);
		if(response.ok)
			return await response.json();
		else
			throw Error("failed to load users: " + response.status);
	}

	async function addUser(mail, password, admin, viewPrivate) {
		const response = await fetch(`${apiUrl}/users`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ mail: mail, password: password, admin: admin, viewPrivate: viewPrivate }),
			credentials: 'include' // Include cookies in subsequent requests
		});

		if(response.ok) {
			location.reload();
		} else {
			if(response.status === 401)
				return location.href = "/login";
			throw new Error("failed to add page: " + response.status);
		}
	}

	async function deleteUser(id) {
		const response = await fetch(`${apiUrl}/users/${id}`, {
			method: 'DELETE',
			credentials: 'include' // Include cookies in subsequent requests
		});

		if(response.ok) {
			location.reload();
		} else {
			if(response.status === 401)
				return location.href = "/login";
			throw new Error("failed to delete user: " + response.status);
		}
	}

	async function getPages() {
		const response = await fetch(`${apiUrl}/pages`);

		const stringResponse = await response.text();
		if(stringResponse.startsWith("]"))
			return [];
		return JSON.parse(stringResponse);
	}

	async function addPage(id, title, description, url, privateAccess) {
		const response = await fetch(`${apiUrl}/pages`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ id: id, title: title, description: description, url: url, privateAccess: privateAccess }),
			credentials: 'include' // Include cookies in subsequent requests
		});

		if(response.ok) {
			location.reload();
		} else {
			if(response.status === 401)
				return location.href = "/login?origin=/admin";
			throw new Error("failed to add page: " + response.status);
		}
	}

	async function deletePage(id) {
		const response = await fetch(`${apiUrl}/pages/${id}`, {
			method: 'DELETE',
			credentials: 'include' // Include cookies in subsequent requests
		});

		if(response.ok) {
			location.reload();
		} else {
			if(response.status === 401)
				return location.href = "/login";
			throw new Error("failed to delete page: " + response.status);
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
		<h3>Add Page</h3>
		<form class="creation">
			<div class="labeled-input">
				<label for="page-id">Id</label>
				<input id="page-id" class="input" type="text" placeholder="id" bind:value={pageToCreateId}/>
			</div>
			<div class="labeled-input">
				<label for="page-title">Title</label>
				<input id="page-title" class="input" type="text" placeholder="title" bind:value={pageToCreateTitle}/>
			</div>

			<div class="labeled-input">
				<label for="page-desc">Description</label>
				<input id="page-desc" class="input" type="text" placeholder="description" bind:value={pageToCreateDesc}/>
			</div>

			<div class="labeled-input">
				<label for="page-url">Url</label>
				<input id="page-url" class="input" type="text" placeholder="/url" bind:value={pageToCreateUrl}/>
			</div>

			<div class="labeled-input">
				<label for="page-private">Private</label>
				<input id="page-private" type="checkbox" class="checkbox" bind:checked={pageToCreatePrivate}/>
			</div>

			<span></span>
			<button type="submit" class="button" on:click|preventDefault={() => addPage(pageToCreateId, pageToCreateTitle, pageToCreateDesc, pageToCreateUrl, pageToCreatePrivate)} disabled={pageToCreateInvalid}>Add</button>
		</form>
		{#await getPages()}
			<p>Loading pages...</p>
		{:then pages }
			<div class="list">
				{#each pages as page}
					<div class="item">
						<p class="cell">{page.id}</p>
						<p class="cell remove-on-small">{page.title}</p>
						<p class="cell remove-on-small">{page.url}</p>
						<p class="cell remove-on-medium">{page.description}</p>

						{#if page.private_access}
							<span class="material-symbols-rounded cell">encrypted</span>
						{:else }
							<span class="material-symbols-rounded cell">lock_open</span>
						{/if}

						<button type="button" class="icon-button" on:click|preventDefault={() => pageIdToDelete = pageIdToDelete === page.id ? undefined : page.id}>
							<span class="material-symbols-rounded">delete</span>
						</button>
					</div>
					{#if page.id === pageIdToDelete}
						<div transition:fly="{{delay: 0, duration: 300, y: -20 }}" class="delete-confirmation">
							<p>Do you really want to delete "{page.url}"?</p>
							<button type="button" class="button" style="filter: opacity(75%)" on:click|preventDefault={() => deletePage(page.id)}>Yes</button>
							<button type="button" class="button" on:click|preventDefault={() => pageIdToDelete = undefined}>Cancel</button>
						</div>
					{/if}
				{/each}
			</div>
		{:catch error }
			<p>{error}</p>
		{/await}

		<h2>Users</h2>
		<h3>Add User</h3>
		<form class="creation">
			<div class="labeled-input">
				<label for="user-mail">Mail</label>
				<input id="user-mail" class="input" type="text" placeholder="mail@demo.com" bind:value={userToCreateMail} autocomplete="username"/>
			</div>

			<div class="labeled-input">
				<label for="user-password">Password</label>
				<input id="user-password" class="input" type="password" placeholder="password" bind:value={userToCreatePassword} autocomplete="new-password"/>
			</div>

			<div class="labeled-input">
				<label for="user-admin">Admin</label>
				<input id="user-admin" type="checkbox" class="checkbox" bind:checked={userToCreateAdmin}/>
			</div>

			<div class="labeled-input">
				<label for="user-private">View Private</label>
				<input id="user-private" type="checkbox" class="checkbox" bind:checked={userToCreatePrivate}/>
			</div>

			<span></span>
			<button type="submit" class="button" on:click|preventDefault={() => addUser(userToCreateMail, userToCreatePassword, userToCreateAdmin, userToCreatePrivate)} disabled={userToCreateInvalid}>Add</button>
		</form>
		{#await getUsers()}
			<p>Loading pages...</p>
		{:then users }
			<div class="list">
				{#each users as user}
					<div class="item">
						<p class="cell remove-on-small">{user.mail}</p>
						{#if user.view_private}
							<span class="material-symbols-rounded cell">lock_open</span>
						{:else }
							<span class="material-symbols-rounded cell">encrypted</span>
						{/if}
						{#if user.admin }
							<span class="material-symbols-rounded cell">admin_panel_settings</span>
						{/if}
						{#if users.length > 1}
							<button type="button" class="icon-button" on:click|preventDefault={() => userIdToDelete = userIdToDelete === user.id ? undefined : user.id}>
								<span class="material-symbols-rounded">delete</span>
							</button>
						{/if}
					</div>
					{#if user.id === userIdToDelete}
						<div transition:fly="{{delay: 0, duration: 300, y: -20 }}" class="delete-confirmation">
							<p>Do you really want to delete "{user.mail}"?</p>
							<button type="button" class="button" style="filter: opacity(75%)" on:click|preventDefault={() => deleteUser(userIdToDelete)}>Yes</button>
							<button type="button" class="button" on:click|preventDefault={() => userIdToDelete = undefined}>Cancel</button>
						</div>
					{/if}
				{/each}
			</div>
		{:catch error }
			<p>{error}</p>
		{/await}
	{:else }
		<p>Not enough permissions!</p>
	{/if}
</main>


<style>
	.list {
		display: flex;
		flex-flow: column;
		gap: 10px;
	}

	.item {
		background-color: var(--michu-tech-primary);
		border: 2px solid var(--michu-tech-accent);
		border-radius: 5px;

		display: flex;
		flex-flow: row nowrap;
		gap: 10px;
		padding: 10px;
		justify-content: space-between;
		align-items: center;
	}

	.item:hover {
		filter: brightness(110%);
	}

	.cell {
		white-space: nowrap;
		text-overflow: ellipsis;
		overflow: hidden;
	}

	.icon-button {
		all: unset;
		border-radius: 50%;
		aspect-ratio: 1 / 1;
		width: 32px;
		display: flex;
		justify-content: center;
		align-items: center;
		cursor: pointer;
		border: 2px solid var(--michu-tech-accent);
	}

	.icon-button:hover {
		background-color: color-mix(in srgb, var(--michu-tech-accent) 50%, transparent);
	}

	.icon-button span {
		color: color-mix(in srgb, var(--michu-tech-primary) 75%, transparent);
	}

	.delete-confirmation {
		display: flex;
		flex-flow: row nowrap;
		justify-content: space-evenly;
		align-items: center;
		gap: 5px;
	}

	.creation {
		margin-bottom: 5vh;
		display: flex;
		flex-flow: row wrap;
		gap: 10px;
		justify-content: space-between;
		align-items: center;
	}

	.creation div {
		min-width: 100px;
	}

	@media (max-width: 1000px) {
		.item .remove-on-medium {
			display: none !important;
		}
	}

	@media (max-width: 600px) {
		.item .remove-on-small {
			display: none !important;
		}
	}
</style>