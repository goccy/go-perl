# app.psgi - a Mojolicious::Lite application served as a PSGI app.
# Loaded by psgi.pl via Plack::Util::load_psgi; app->start('psgi') returns
# the PSGI code ref.
use Mojolicious::Lite -signatures;

get '/' => sub ($c) {
    $c->render(text => "Hello from Mojolicious running on go-perl!\n");
};

get '/json' => sub ($c) {
    $c->render(json => {
        framework => 'Mojolicious',
        runtime   => 'go-perl',
        perl      => "$^V",
    });
};

post '/echo' => sub ($c) {
    $c->render(text => uc $c->req->body);
};

# Surfaces the PSGI environment the server hands the app (psgi.* flags).
get '/psgi-flags' => sub ($c) {
    my $env = $c->req->env;
    $c->render(text => join ',',
        'multiprocess=' . ($env->{'psgi.multiprocess'} ? 1 : 0),
        'multithread=' . ($env->{'psgi.multithread'} ? 1 : 0),
        'streaming=' . ($env->{'psgi.streaming'} ? 1 : 0));
};

# Served through a real XS module: Cpanel::JSON::XS here is the stock CPAN
# distribution compiled against go-perl's native XS SDK (`gperl xs build`).
use Cpanel::JSON::XS ();
my $xs_json = Cpanel::JSON::XS->new->canonical->utf8;
get '/fast-json' => sub ($c) {
    $c->render(data => $xs_json->encode({
        encoder => 'Cpanel::JSON::XS',
        native  => \1,
        values  => [1 .. 5],
    }), format => 'json');
};

app->start('psgi');
