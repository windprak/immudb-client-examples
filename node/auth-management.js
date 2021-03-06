/*
Copyright 2019-2020 vChain, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

const ImmudbClient = require('immudb-node')
const types = require('immudb-node/lib/types')

const IMMUDB_HOST = '127.0.0.1'
const IMMUDB_PORT = 3322
const IMMUDB_USER = 'immudb'
const IMMUDB_PWD = 'immudb'

ImmudbClient({
  address: `${IMMUDB_HOST}:${IMMUDB_PORT}`,
}, main)

async function main(err, cl) {
  if (err) {
    return console.log(err)
  }

  try {
    let req = { username: IMMUDB_USER, password: IMMUDB_PWD }
    let res = await cl.login(req)

    res = await cl.useDatabase({ database: 'defaultdb' })
    console.log('success: useDatabase:', res)

    res = await cl.updateAuthConfig({ auth: types.auth.enabled })
    console.log('success: updateAuthConfig')

    res = await cl.updateMTLSConfig({ enabled: false })
    console.log('success: updateMTLSConfig')

  } catch (err) {
    console.log(err)
  }
}
