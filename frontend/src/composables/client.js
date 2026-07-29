import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { FaridoonService } from '../../gen/faridoon/v1/faridoon_pb.ts'

const transport = createConnectTransport({
  baseUrl: '',
  credentials: 'include',
})

export const client = createClient(FaridoonService, transport)
