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

app->start('psgi');
