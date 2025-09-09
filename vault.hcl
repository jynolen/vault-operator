ui            = true
cluster_addr  = "http://127.0.0.1:8201"
api_addr      = "http://127.0.0.1:8200"
disable_mlock = true

storage "inmem" {
}

listener "tcp" {
    tls_disable = true
}

listener "unix" {
  address = "/tmp/vault.sock"
}

telemetry {
  statsite_address = "127.0.0.1:8125"
  disable_hostname = true
}