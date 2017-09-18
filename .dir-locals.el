((go-mode
  . ((go-test-args . "-tags libsqlite3 -timeout 5s")
     (eval
      . (set
	 (make-local-variable 'flycheck-go-build-tags)
	 '("libsqlite3"))))))
