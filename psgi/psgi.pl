# psgi.pl - the PSGI adapter loaded once into every pooled interpreter.
#
# psgi_load compiles the PSGI application (app.psgi) via Plack::Util.
# psgi_handle receives one HTTP request from Go - the CGI-style environment
# as a hash ref and the request body base64-encoded - runs the app, and
# returns (status, header pairs, base64 body). The body crosses the Go<->Perl
# boundary base64-encoded in BOTH directions because the bridge speaks JSON,
# and JSON strings cannot carry arbitrary bytes.
use strict;
use warnings;
use MIME::Base64 qw(decode_base64 encode_base64);
use Plack::Util;

my $app;

sub psgi_load {
    my ($path) = @_;
    $app = Plack::Util::load_psgi($path);
    die "$path did not return a PSGI application\n" unless ref $app eq 'CODE';
    return 1;
}

sub psgi_handle {
    my ($env, $body_b64) = @_;
    die "psgi_load has not been called\n" unless $app;

    my $body = decode_base64(defined $body_b64 ? $body_b64 : '');
    open my $in, '<', \$body or die "open request body: $!\n";

    $env->{'psgi.version'}      = [1, 1];
    $env->{'psgi.url_scheme'}   = $env->{'psgi.url_scheme'} || 'http';
    $env->{'psgi.input'}        = $in;
    $env->{'psgi.errors'}       = \*STDERR;
    $env->{'psgi.multithread'}  = Plack::Util::FALSE;
    # the server passes psgi.multiprocess when other workers serve
    # concurrently; a lone instance defaults to false
    $env->{'psgi.multiprocess'} = Plack::Util::FALSE
        unless exists $env->{'psgi.multiprocess'};
    $env->{'psgi.run_once'}     = Plack::Util::FALSE;
    $env->{'psgi.streaming'}    = Plack::Util::FALSE;
    $env->{'psgi.nonblocking'}  = Plack::Util::FALSE;

    my $res = $app->($env);
    die "PSGI app returned a non-arrayref response (psgi.streaming is off)\n"
        unless ref $res eq 'ARRAY';
    my ($status, $headers, $out_body) = @$res;

    my $out = '';
    Plack::Util::foreach($out_body, sub { $out .= $_[0] if defined $_[0] });
    if (ref($out_body) ne 'ARRAY') {
        eval { $out_body->close };
    }
    # encode_base64 wants octets; a character string (wide chars) would die.
    utf8::encode($out) if utf8::is_utf8($out);

    return ($status, $headers, encode_base64($out, ''));
}

1;
