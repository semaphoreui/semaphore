define(function () {
    app.registerController('ProjectRepositoriesCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', function ($scope, $http, Project, $modal, $rootScope) {
        $scope.reload = function () {
            $http.get(Project.getURL() + '/keys?type=ssh&sort=name&order=asc').then(function (keys) {
                $scope.sshKeys = keys.data;

                $http.get(Project.getURL() + '/repositories?sort=name&order=asc').then(function (repos) {
                    repos.data.forEach(function (repo) {
                        for (var i = 0; i < keys.length; i++) {
                            if (repo.ssh_key_id == keys[i].id) {
                                repo.ssh_key = keys[i];
                                break;
                            }
                        }
                    });

                    $scope.repositories = repos.data;
                });
            });
        }

        $scope.remove = function (repo) {
            $http.delete(Project.getURL() + '/repositories/' + repo.id)
                .then(function () {
                    $scope.reload();
                })
                .catch(function (response) {
                    var d = response.data;
                    if (!(d && d.templatesUse)) {
                        swal('error', 'could not delete repository..', 'error');
                        return;
                    }

                    swal({
                        title: 'Repository in use',
                        text: d.error,
                        icon: 'error',
                        buttons: {
                            cancel: true,
                            confirm: {
                                text: 'Mark as removed',
                                closeModel: false,
                                className: 'bg-danger',
                            }
                        }
                    }).then(function (value) {
                        if (!value) {
                            return;
                        }

                        $http.delete(Project.getURL() + '/repositories/' + repo.id + '?setRemoved=1')
                            .then(function () {
                                swal.stopLoading();
                                swal.close();

                                $scope.reload();
                            })
                            .catch(function () {
                                swal('Error', 'Could not delete repository..', 'error');
                            });
                    });
                });
        }

        $scope.update = function (repo) {
            var scope = $rootScope.$new();
            scope.keys = $scope.sshKeys;
            scope.repo = JSON.parse(JSON.stringify(repo));

            $modal.open({
                templateUrl: '/tpl/projects/repositories/add.html',
                scope: scope
            }).result.then(function (opts) {
                if (opts.remove) {
                    return $scope.remove(repo);
                }

                $http.put(Project.getURL() + '/repositories/' + repo.id, opts.repo).then(function () {
                    $scope.reload();
                }).catch(function (response) {
                    swal('Error', 'Repository not updated: ' + response.status, 'error');
                });
            }, function () {
            });
        }

        $scope.add = function () {
            var scope = $rootScope.$new();
            scope.keys = $scope.sshKeys;

            $modal.open({
                templateUrl: '/tpl/projects/repositories/add.html',
                scope: scope
            }).result.then(function (repo) {
                $http.post(Project.getURL() + '/repositories', repo.repo)
                    .then(function () {
                        $scope.reload();
                    }).catch(function (response) {
                    swal('Error', 'Repository not added: ' + response.status, 'error');
                });
            }, function () {
            });
        }

        $scope.reload();
    }]);
});
