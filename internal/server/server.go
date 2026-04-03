package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-presskit/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/assets",s.listAssets);s.mux.HandleFunc("POST /api/assets",s.createAsset);s.mux.HandleFunc("DELETE /api/assets/{id}",s.deleteAsset)
s.mux.HandleFunc("GET /api/facts",s.listFacts);s.mux.HandleFunc("POST /api/facts",s.createFact);s.mux.HandleFunc("DELETE /api/facts/{id}",s.deleteFact)
s.mux.HandleFunc("GET /api/contacts",s.listContacts);s.mux.HandleFunc("POST /api/contacts",s.createContact);s.mux.HandleFunc("DELETE /api/contacts/{id}",s.deleteContact)
s.mux.HandleFunc("GET /api/kit",s.publicKit)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/presskit/"})})
return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listAssets(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"assets":oe(s.db.ListAssets())})}
func(s *Server)createAsset(w http.ResponseWriter,r *http.Request){var a store.Asset;json.NewDecoder(r.Body).Decode(&a);if a.Name==""{we(w,400,"name required");return};a.Public=true;s.db.CreateAsset(&a);wj(w,201,a)}
func(s *Server)deleteAsset(w http.ResponseWriter,r *http.Request){s.db.DeleteAsset(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listFacts(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"facts":oe(s.db.ListFacts())})}
func(s *Server)createFact(w http.ResponseWriter,r *http.Request){var f store.Fact;json.NewDecoder(r.Body).Decode(&f);if f.Label==""{we(w,400,"label required");return};s.db.CreateFact(&f);wj(w,201,f)}
func(s *Server)deleteFact(w http.ResponseWriter,r *http.Request){s.db.DeleteFact(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listContacts(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"contacts":oe(s.db.ListContacts())})}
func(s *Server)createContact(w http.ResponseWriter,r *http.Request){var c store.Contact;json.NewDecoder(r.Body).Decode(&c);if c.Name==""{we(w,400,"name required");return};s.db.CreateContact(&c);wj(w,201,c)}
func(s *Server)deleteContact(w http.ResponseWriter,r *http.Request){s.db.DeleteContact(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)publicKit(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.PublicKit())}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"presskit","assets":st.Assets})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
