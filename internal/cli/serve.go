package cli

import "devia/internal/server"

func cmdServe(args []string) {
	fs := newFlagSet("serve")
	port := fs.Int("port", 7654, "TCP port to listen on")
	host := fs.String("host", "127.0.0.1", "address to bind (127.0.0.1 = local only)")
	parseArgs(fs, args)

	if err := server.Run(*host, *port); err != nil {
		fail(err)
	}
}
