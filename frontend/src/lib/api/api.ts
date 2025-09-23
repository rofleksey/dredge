import axios from 'axios'
import { computed, onBeforeMount, onBeforeUnmount } from 'vue'
import PQueue from 'p-queue'
import {DefaultApiFactory, type DefaultApiInterface} from "../oapi";
import {useAuthStore} from "../../stores/auth-store.ts";


// @ts-ignore
const BASE_URL = process.env.NODE_ENV === 'development' ? "http://localhost:8080/v1" : "/v1"

export interface ApiClientOptions {
  concurrency?: number,
  intervalCap?: number,
  interval?: number,
}

export class ApiClient {
  public api: DefaultApiInterface

  public isDestroyed = false
  private controller: AbortController

  constructor (opts?: ApiClientOptions) {
    const controller = new AbortController()

    const authStore = useAuthStore()
    const token = computed(() => authStore.token)

    const instance = axios.create({
      baseURL: BASE_URL,
      timeout: 10000,
      signal: controller.signal,
    })

    instance.interceptors.request.use((cfg) => {
      if (token.value) {
        cfg.headers.set('Authorization', `Bearer ${token.value}`)
      }
      return cfg
    })

    const api  = {
      ...DefaultApiFactory(undefined, "", instance)
    }

    const queue = new PQueue({
      concurrency: opts?.concurrency ?? 1,
      intervalCap: opts?.intervalCap ?? Infinity,
      interval: opts?.interval ?? 0,
      throwOnTimeout: true
    })

    // decorate all functions to use queue
    for (const key in api) {
      // @ts-ignore
      const original = api[key]

      if (typeof original !== 'function') {
        continue
      }

      // @ts-ignore
      api[key] = function() {
        return queue.add(() => original.apply(api, arguments))
      }
    }

    this.api = api
    this.controller = controller
  }

  destroy () {
    this.isDestroyed = true
    this.controller.abort()
  }
}

export function useApiClient (opts?: ApiClientOptions): ApiClient {
  let client = new ApiClient(opts)

  onBeforeMount(() => {
    if (client.isDestroyed) {
      client = new ApiClient(opts)
    }
  })

  onBeforeUnmount(() => {
    client.destroy()
  })

  return client
}
