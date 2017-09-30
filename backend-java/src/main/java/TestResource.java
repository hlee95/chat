import java.util.List;
import java.util.Map;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import org.skife.jdbi.v2.DBI;

@Path("/test")
public class TestResource {
    private final DBI dbi;

    public TestResource(DBI dbi) {
        this.dbi = dbi;
    }

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    public TestRepresentation get() {
        return this.dbi.withHandle((handle) -> {
            List<Map<String, Object>> rs = handle.select("SELECT col FROM test");
            TestRepresentation repr = new TestRepresentation();
            repr.result = (String) rs.get(0).get("col");
            repr.backend = "java";
            return repr;
        });
    }
}
