import io.dropwizard.Configuration;
import io.dropwizard.db.DataSourceFactory;

class ChallengeConfiguration extends Configuration {
    public DataSourceFactory database = new DataSourceFactory();
}
