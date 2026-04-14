import { isbot } from "isbot"
import { renderToPipeableStream } from "react-dom/server"
import { ServerRouter } from "react-router"

import type { EntryContext } from "react-router"

const ABORT_DELAY = 5_000

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  routerContext: EntryContext,
) {
  return new Promise((resolve, reject) => {
    let shellRendered = false
    const userAgent = request.headers.get("user-agent")
    const callbackName = isbot(userAgent ?? "") ? "onAllReady" : "onShellReady"

    const { pipe, abort } = renderToPipeableStream(
      <ServerRouter context={routerContext} url={request.url} />,
      {
        [callbackName]() {
          shellRendered = true

          const body = new ReadableStream({
            start(controller) {
              const encoder = new TextEncoder()

              const writable = new WritableStream({
                write(chunk: Uint8Array) {
                  controller.enqueue(chunk)
                },
                close() {
                  controller.close()
                },
                abort(reason) {
                  controller.error(reason)
                },
              })

              const writer = writable.getWriter()

              pipe({
                write(chunk: string | Uint8Array) {
                  const bytes =
                    typeof chunk === "string" ? encoder.encode(chunk) : chunk
                  writer.write(bytes)
                  return true
                },
                end() {
                  writer.close()
                },
                on() {},
                off() {},
                removeListener() {},
              } as unknown as NodeJS.WritableStream)
            },
          })

          responseHeaders.set("Content-Type", "text/html")

          resolve(
            new Response(body, {
              headers: responseHeaders,
              status: responseStatusCode,
            }),
          )
        },
        onShellError(error: unknown) {
          reject(error)
        },
        onError(error: unknown) {
          if (shellRendered) {
            console.error(error)
          }
          responseStatusCode = 500
        },
      },
    )

    setTimeout(abort, ABORT_DELAY)
  })
}
