# MCP inventory

Run the cases in order. Cases 1–3 share the inventory **Tools**, cases 5–8 share **Rejects** and **Other**; cases 4 and 9 create nothing. The letters stand for IDs recorded during the run: **A** — Tools, **B** — Rejects, **O** — Other, **H** — Hosts, **G**, **D** and **F** — MCP servers. To read an inventory, send `GET /api/project/<project ID>/inventory/<inventory ID>` with an authenticated HTTP client signed in as the user the case names; `GET /api/project/<project ID>/inventory` lists the inventories with their IDs. An API case sends JSON with content type `application/json`.

## Fixture

Use the English UI on a stand built from the PR branch, any edition. Switch on **Claude Code Agent** and **Codex Agent** on the **Applications** page; the suite never starts an agent, so neither command line tool has to be installed. Sign in as an administrator, who becomes the owner of the project.

1. Create a project named **MCP QA**. Its **Key Store** holds one key, **None**, and it has no inventories.
2. Create two users who are not administrators and add them to the **Team** of the project.

   | Username | Password | Project role |
   | --- | --- | --- |
   | `e2e-guest` | `e2e-password-1` | **Guest** |
   | `e2e-runner` | `e2e-password-1` | **Task Runner** |

3. Create an Ansible inventory named **Hosts** of type **Static** with the content `localhost`. Record its ID as **H**.

The token values of this suite are `e2e-token-github`, `e2e-token-search`, `e2e-token-search-2`, `e2e-token-docs` and `e2e-token-foreign`. Wherever a case says that a response holds no token, search the whole response body for each of them.

## 1. The New Inventory menu creates an MCP inventory with its servers

1. Open **Inventory** and click **New Inventory**. The menu lists **Ansible Inventory** and **MCP Inventory**.
2. Click **MCP Inventory**. The dialog **New MCP Inventory** opens with the field **Name**, the section **MCP servers** and one empty server row: the switch is on, the placeholders are **Name**, `https://mcp.example.com/mcp` and **Token**.
3. Enter the name `Tools`. Add two rows with the **+** button (its tooltip is **Add server**) and fill the three rows.

   | Switch | Name | URL | Token |
   | --- | --- | --- | --- |
   | on | `github` | `https://mcp.example.com/github` | `e2e-token-github` |
   | on | `docs` | `https://mcp.example.com/docs` | empty |
   | off | `search` | `https://mcp.example.com/search` | `e2e-token-search` |

   A typed token is shown as dots.
4. Click **Create**. The dialog closes and the list shows **Tools** with the type `mcp`. Record its ID as **A**.
5. Read A as the owner. `type` is `mcp`, `inventory` is `""`, and `mcp_servers` holds the three servers in the order of the table with these values of `enabled` and `has_token`: `github` — `true`, `true`; `docs` — `true`, `false`; `search` — `false`, `true`. The response holds no token and no `token` field. Record the `id` of each server.
6. Read the list of inventories. The entry of Tools has no `mcp_servers` field.
7. Open **Key Store**. It still lists only **None**.

## 2. A saved token is never shown, and can be removed, added and replaced

1. On **Inventory** click the pencil of **Tools**. The dialog **Edit MCP Inventory** opens. Every token field is empty. The placeholder is `*****` for `github` and `search`, and **Token** for `docs`. Only `github` and `search` have a cross in the token field; its tooltip is **Remove token**.
2. Click the cross of `github`. The placeholder becomes **Will be removed** and the tooltip of the button becomes **Keep token**.
3. Type `e2e-token-docs` into the token field of `docs` and `e2e-token-search-2` into the one of `search`. Click **Save**.
4. Read A. `has_token` is `false` for `github` and `true` for `docs` and `search`. The `id`, `name`, `url` and `enabled` of every server are as recorded in case 1. The response holds no token.
5. Open the dialog again. The placeholder is **Token** for `github`, which has no cross now, and `*****` for `docs` and `search`.
6. Click the cross of `docs` twice. After the first click the placeholder is **Will be removed**; after the second it is `*****` again and the tooltip is **Remove token**. Click **Cancel**.
7. Read A. The response equals the one of step 4.

## 3. Servers are switched off, renamed, removed and added in one save

1. Open the dialog of **Tools**. Switch `docs` off. Rename `search` to `search-v2`. Click the bin of `github` (its tooltip is **Remove server**). Add a row with the name `wiki`, the URL `http://wiki.internal:8080/mcp` and no token. Click **Save**.
2. Read A. It has three servers.

   | `name` | `id` | `enabled` | `has_token` |
   | --- | --- | --- | --- |
   | `docs` | as recorded for `docs` | `false` | `true` |
   | `search-v2` | as recorded for `search` | `false` | `true` |
   | `wiki` | a new one | `true` | `false` |

   There is no `github`.
3. Open the dialog again and remove every row. The table shows **No servers. The agent runs without MCP tools.** Click **Save**.
4. Read A. It has no `mcp_servers` field. Open the dialog once more: it shows no server row and the same text. Click **Cancel**.

## 4. The form refuses an invalid server without sending it

1. Click **New Inventory → MCP Inventory** and click **Create** on the untouched form. **Name is required** appears under the name of the inventory and under the name of the server row, and **URL is required** under its URL. The dialog stays open.
2. Enter the inventory name `Form checks`. Add a second row with the name `github` and the URL `https://mcp.example.com/github`.
3. For each line of the table, set the name and the URL of the first row and click **Create**. The message appears under the named field, and the dialog stays open.

   | Name | URL | Message | Under |
   | --- | --- | --- | --- |
   | empty | `https://mcp.example.com/a` | **Name is required** | the name |
   | `my server` | `https://mcp.example.com/a` | **Up to 64 letters, digits, - and _** | the name |
   | the letter `a` 65 times | `https://mcp.example.com/a` | **Up to 64 letters, digits, - and _** | the name |
   | `github` | `https://mcp.example.com/a` | **Another server has this name** | the name of both rows |
   | `ok` | empty | **URL is required** | the URL |
   | `ok` | `ftp://mcp.example.com/a` | **An http or https URL without a user name or password** | the URL |
   | `ok` | `https://user:pass@mcp.example.com/a` | **An http or https URL without a user name or password** | the URL |
   | `ok` | `mcp.example.com/a` | **An http or https URL without a user name or password** | the URL |

4. Click **Cancel**. The list of inventories has no **Form checks**.

## 5. The API rejects an invalid MCP inventory and stores nothing of it

As the owner, create two inventories with `POST /api/project/<project ID>/inventory`.

- `{"project_id": <project ID>, "type": "mcp", "name": "Rejects", "mcp_servers": [{"name": "github", "url": "https://mcp.example.com/github", "enabled": true, "token": {"operation": "create", "secret": "e2e-token-github"}}, {"name": "docs", "url": "https://mcp.example.com/docs", "enabled": true}]}` — the response is 201 and lists both servers with `has_token` `true` and `false` and no token. Record the inventory ID as **B** and the server IDs as **G** (`github`) and **D** (`docs`).
- `{"project_id": <project ID>, "type": "mcp", "name": "Other", "mcp_servers": [{"name": "foreign", "url": "https://mcp.example.com/foreign", "enabled": true, "token": {"operation": "create", "secret": "e2e-token-foreign"}}]}` — record the inventory ID as **O** and the server ID as **F**.

1. Record the names in the list of inventories. For each line, POST `{"project_id": <project ID>, "type": "mcp", "name": "Rejected"}` with the listed fields added or replaced. In a server, `…` stands for `"name": "github", "url": "https://mcp.example.com/github", "enabled": true`, with the listed field replacing its own. Every response is HTTP 400 with the given JSON `error`, and the list of inventories has the same names after it.

   | Added to the body | `error` |
   | --- | --- |
   | `"mcp_servers": [{…, "name": "my server"}]` | `mcp server #1: name may contain only letters, digits, '-' and '_'` |
   | `"mcp_servers": [{…, "name": ""}]` | `mcp server #1: name can not be empty` |
   | `"mcp_servers": [{…, "name": <the letter a 65 times>}]` | `mcp server #1: name can not be longer than 64 characters` |
   | `"mcp_servers": [{…}, {…, "url": "https://mcp.example.com/other"}]` | `mcp server name is not unique: github` |
   | `"mcp_servers": [{…, "url": ""}]` | `mcp server #1: url can not be empty` |
   | `"mcp_servers": [{…, "url": "ftp://mcp.example.com/x"}]` | `mcp server #1: url must use http or https` |
   | `"mcp_servers": [{…, "url": "mcp.example.com/x"}]` | `mcp server #1: url must use http or https` |
   | `"mcp_servers": [{…, "url": "https://"}]` | `mcp server #1: url must have a host` |
   | `"mcp_servers": [{…, "url": "https://user:pass@mcp.example.com/x"}]` | `mcp server #1: url must not contain credentials, use a token instead` |
   | `"mcp_servers": [{…}, {…, "name": "docs", "url": "ftp://mcp.example.com/x"}]` | `mcp server #2: url must use http or https` |
   | `"mcp_servers": [{…, "token": {"operation": "create"}}]` | `mcp server github: token secret can not be empty` |
   | `"mcp_servers": [{…, "token": {"operation": "rotate", "secret": "x"}}]` | `mcp server github: invalid token operation: rotate` |
   | `"mcp_servers": [{…, "token": {"operation": "update", "secret": "x"}}]` | `mcp server github has no token, use create` |
   | `"inventory": "localhost"` | `mcp inventory must not have the inventory field, use mcp_servers` |
   | `"ssh_key_id": <ID of the key None>` | `mcp inventory must not have ssh_key_id or become_key_id` |
   | `"become_key_id": <ID of the key None>` | `mcp inventory must not have ssh_key_id or become_key_id` |
   | `"type": "static", "inventory": "localhost", "mcp_servers": [{…}]` | `mcp_servers is allowed only for an mcp inventory` |

2. Read B and keep the response. For each line, PUT `{"id": <B>, "project_id": <project ID>, "type": "mcp", "name": "Rejects"}` to `/api/project/<project ID>/inventory/<B>` with the listed fields added or replaced. `g` stands for `"id": <G>, "name": "github", "url": "https://mcp.example.com/github", "enabled": true` and `d` for `"id": <D>, "name": "docs", "url": "https://mcp.example.com/docs", "enabled": true`. Every response is HTTP 400 with the given `error`, and B read again equals the kept response.

   | Added to the body | `error` |
   | --- | --- |
   | `"mcp_servers": [{g, "token": {"operation": "create", "secret": "x"}}, {d}]` | `mcp server github already has a token, use update` |
   | `"mcp_servers": [{g}, {d, "token": {"operation": "update", "secret": "x"}}]` | `mcp server docs has no token, use create` |
   | `"mcp_servers": [{g}, {d}, {"id": <F>, "name": "foreign", "url": "https://mcp.example.com/foreign", "enabled": true}]` | `mcp server <F> does not belong to the inventory` |
   | `"mcp_servers": [{g}, {g, "name": "github2"}, {d}]` | `mcp server <G> is listed twice` |
   | `"mcp_servers": [{g}, {d, "name": "github"}]` | `mcp server name is not unique: github` |
   | `"mcp_servers": [{g, "url": "ftp://mcp.example.com/x"}, {d}]` | `mcp server #1: url must use http or https` |
   | `"inventory": "localhost"` | `mcp inventory must not have the inventory field, use mcp_servers` |
   | `"type": "static", "inventory": "localhost"` | `inventory type can not be changed to or from mcp` |
   | `"id": <O>` | `Inventory ID in body and URL must be the same` |

3. PUT `{"id": <H>, "project_id": <project ID>, "type": "mcp", "name": "Hosts"}` to `/api/project/<project ID>/inventory/<H>`. The response is HTTP 400 with `inventory type can not be changed to or from mcp`. H read again still has the type `static` and the content `localhost`.
4. PUT `{"id": <H>, "project_id": <project ID>, "type": "static", "name": "Hosts", "inventory": "localhost", "mcp_servers": [{"name": "github", "url": "https://mcp.example.com/github", "enabled": true}]}` to the same URL. The response is HTTP 400 with `mcp_servers is allowed only for an mcp inventory`.
5. POST `{"name": "sneaky", "type": "none", "project_id": <project ID>, "owner": "mcp"}` to `/api/project/<project ID>/keys`. The response is HTTP 400 with `Invalid key owner`, and `GET /api/project/<project ID>/keys` still returns only the key **None**.

## 6. An update without mcp_servers keeps the servers, an empty list removes them

1. As the owner, PUT `{"id": <B>, "project_id": <project ID>, "type": "mcp", "name": "Rejects renamed"}` to `/api/project/<project ID>/inventory/<B>`. The response is HTTP 204. B read again has the name `Rejects renamed` and both servers with the IDs G and D and the same `has_token` as before.
2. PUT the same body with `"mcp_servers": []` added. The response is HTTP 204. B read again has no `mcp_servers` field.
3. PUT the same body with `"mcp_servers": [{"name": "github", "url": "https://mcp.example.com/github", "enabled": true, "token": {"operation": "create", "secret": "e2e-token-github"}}]`. The response is HTTP 204. B read again has one server `github` with `has_token` `true` and an `id` that is not G.

## 7. A member who can not manage resources reads the servers without tokens and changes nothing

Read O as the owner and keep the response. Run the steps as `e2e-guest`, then again as `e2e-runner`, each in a separate browser session.

1. Open **Inventory**. The list shows **Hosts**, **Other**, **Rejects renamed** and **Tools**. There is no **New Inventory** button, and no row has a pencil or a bin.
2. Read O. The response is HTTP 200 with the server `foreign` and `has_token` `true`, and holds no token.
3. Send three requests. Each response is HTTP 403.

   | Request | Body |
   | --- | --- |
   | `POST /api/project/<project ID>/inventory` | `{"project_id": <project ID>, "type": "mcp", "name": "By member", "mcp_servers": [{"name": "github", "url": "https://mcp.example.com/github", "enabled": true}]}` |
   | `PUT /api/project/<project ID>/inventory/<O>` | `{"id": <O>, "project_id": <project ID>, "type": "mcp", "name": "Other", "mcp_servers": []}` |
   | `DELETE /api/project/<project ID>/inventory/<O>` | none |

After both users: as the owner, O read again equals the kept response, and the list of inventories has no **By member**.

## 8. Deleting an MCP inventory removes it

1. As the owner, open **Inventory** and click the bin of **Other**. The dialog **Delete inventory** opens. Click **Yes**.
2. The list shows **Hosts**, **Rejects renamed** and **Tools**. `GET /api/project/<project ID>/inventory/<O>` returns HTTP 404.

## 9. The menu offers an MCP inventory only while an agent application is switched on

The case changes the applications of the whole stand and puts them back in its last step.

1. As the administrator, open **Applications** and switch **Claude Code Agent** off. Open **Inventory** of the project and click **New Inventory**. The menu still lists **Ansible Inventory** and **MCP Inventory**.
2. On **Applications** switch **Codex Agent** off as well. Open **Inventory** and click **New Inventory**. The menu lists only **Ansible Inventory**. The list still shows **Tools** and **Rejects renamed** with the type `mcp`, and the pencil of **Tools** still opens **Edit MCP Inventory**. Click **Cancel**.
3. On **Applications** switch both applications on. Open **Inventory** and click **New Inventory**. The menu lists **Ansible Inventory** and **MCP Inventory** again.
