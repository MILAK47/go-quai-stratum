package proxy

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/INFURA/go-ethlibs/jsonrpc"

	"github.com/dominant-strategies/go-quai-stratum/util"
	"github.com/dominant-strategies/go-quai/core/types"
)

// Allow only lowercase hexadecimal with 0x prefix
var noncePattern = regexp.MustCompile("^0x[0-9a-f]{16}$")
var hashPattern = regexp.MustCompile("^0x[0-9a-f]{64}$")
var workerPattern = regexp.MustCompile("^[0-9a-zA-Z-_]{1,8}$")

// Clients should provide a Quai address when logging in.
func (s *ProxyServer) handleLoginRPC(cs *Session, params jsonrpc.Params) error {
	if len(params) == 0 {
		return fmt.Errorf("invalid params")
	}

	addy, err := strconv.Unquote(string(params[0]))
	if err != nil {
		log.Printf("%v", err)
	}
	login := strings.ToLower(addy)
	if !util.IsValidHexAddress(login) {
		return fmt.Errorf("invalid login")
	}

	if !s.policy.ApplyLoginPolicy(login, cs.ip) {
		return fmt.Errorf("you are blacklisted")
	}
	cs.login = login
	s.registerSession(cs)
	log.Printf("Stratum miner connected %v@%v", login, cs.ip)

	if s.config.Proxy.Stratum.Enabled {
		go s.broadcastNewJobs()
	}

	return nil
}

// Returns the cached header to clients.
func (s *ProxyServer) handleGetWorkRPC(cs *Session) (*types.Header, *ErrorReply) {
	t := s.currentBlockTemplate()
	if t == nil || t.Header == nil || s.isSick() {
		return nil, &ErrorReply{Code: 0, Message: "Work not ready"}
	}
	return t.Header, nil
}
