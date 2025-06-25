import axios from 'axios';

export async function loadProjectResources(name) {
  return (await axios({
    method: 'get',
    url: `/api/project/${this.projectId}/${name}`,
    responseType: 'json',
  })).data;
}

export async function test() {
  return null;
}
