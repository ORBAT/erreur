// package erreur implements structured errors that allow you to add context fields to errors, with
// fast JSON serialization.
//
// Structured errors
//
// Structured logging with frameworks like zap (https://godoc.org/go.uber.org/zap) is fairly common
// nowadays, but (at least as far as I could tell) errors with contextual information are still
// mainly produced with something like Errorf:
// 	connErr := fmt.Errorf("error code %d when connecting to address %s", errCode, someAddr)
//
// To get all the context into a structured log line, you now need to either log the error while the context is in scope (and depending on what you're writing this might not be sensible), return the context somehow, or make do with what you have:
//
// 	zap.NewExample().Error("failed to load data", zap.Error(connErr), zap.String("addr", someAddr))
// 	Output: {"level":"error","msg":"failed to load data","error":"error code 1234 when connecting to address example.com","addr":"example.com"}
//
// At this point you've lost the benefits of structured logging. So why not have errors that can
// provide their own context? The erreur package provides a zap-specific (although useful outside of
// zap) way of creating structured errors to go along with your structured logging:
//
//  connErr := erreur.New("connection error", zap.Int("code", 1234), zap.String("addr", "example.com"))
//  // [...] elsewhere in your code
//
//  zap.NewExample().Error("failed to load data", erreur.Field(connErr))
//  // Output: {"level":"error","msg":"failed to load data","error":{"msg":"connection error","code":1234,"addr":"example.com"}}
//
//  // or alternatively
//  json := connErr.JSON()
package erreur

