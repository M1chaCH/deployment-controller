export async function getUsers(apiUrl) {
  const response = await fetch(`${apiUrl}/users`);
  if(response.ok)
    return await response.json();
  else
    throw Error("failed to load users: " + response.status);
}

export async function addUser(apiUrl, id, mail, password, admin, pagesToAllow) {
  const response = await fetch(`${apiUrl}/users`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, mail, password, admin, active: false, pagesToAllow, pagesToDisallow: [] }),
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

export async function editUser(apiUrl, id, password, admin, active, pagesToAllow, pagesToDisallow) {
  const response = await fetch(`${apiUrl}/users`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, mail: null, password: password ?? null, admin, active, pagesToAllow, pagesToDisallow }),
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

export async function deleteUser(apiUrl, id) {
  const response = await fetch(`${apiUrl}/users/${id}`, {
    method: 'DELETE',
    credentials: 'include' // Include cookies in subsequent requests
  });

  if(response.ok) {
    location.reload();
  } else {
    if(response.status === 401)
      return location.href = "/login?origin=/admin";
    throw new Error("failed to delete user: " + response.status);
  }
}

export async function getPages(apiUrl) {
  const response = await fetch(`${apiUrl}/pages`);

  const stringResponse = await response.text();
  if(stringResponse.startsWith("]"))
    return [];
  return JSON.parse(stringResponse);
}

export async function addPage(apiUrl, id, url, title, description, privatePage) {
  const response = await fetch(`${apiUrl}/pages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, url, title, description, privatePage }),
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

export async function editPage(apiUrl, id, url, title, description, privatePage) {
  const response = await fetch(`${apiUrl}/pages`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, url, title, description, privatePage }),
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

export async function deletePage(apiUrl, id) {
  const response = await fetch(`${apiUrl}/pages/${id}`, {
    method: 'DELETE',
    credentials: 'include' // Include cookies in subsequent requests
  });

  if(response.ok) {
    location.reload();
  } else {
    if(response.status === 401)
      return location.href = "/login?origin=/admin";
    throw new Error("failed to delete page: " + response.status);
  }
}